package lumi

import (
	"errors"
	fmt "fmt"
	"sync"

	"github.com/rs/zerolog/log"
	"go.mondoo.io/mondoo/falcon"
	"go.mondoo.io/mondoo/motor"
)

// NewResource creates the base class for a new resource
// called during the factory method of every resource creation
func (ctx *Runtime) NewResource(name string) *Resource {
	// initialize resource
	return &Resource{
		MotorRuntime: ctx,
		ResourceID:   ResourceID{Name: name},
	}
}

// NotReadyError indicates the results are not ready to be processed further
type NotReadyError struct{}

func (n NotReadyError) Error() string {
	return "NotReadyError"
}

var NotFound = errors.New("not found")

// CacheEntry contains cached data for resources
type CacheEntry struct {
	Timestamp int64
	Valid     bool
	Data      interface{}
	Error     error
}

// Cache is a map containing CacheEntry values
type Cache struct{ sync.Map }

// Store a new call connection
func (c *Cache) Store(key string, v *CacheEntry) { c.Map.Store(key, v) }

// Load a call connection
func (c *Cache) Load(key string) (*CacheEntry, bool) {
	res, ok := c.Map.Load(key)
	if res == nil {
		return nil, ok
	}
	return res.(*CacheEntry), ok
}

// Delete a Cache Entry
func (c *Cache) Delete(key string) { c.Map.Delete(key) }

// mondoo platform config so that resource scan talk upstream
// TODO: this configuration struct does not belong into the lumi package
// nevertheless the lumi runtime needs to have something that allows users
// to store additional credentials so that resource can use those for
// their resources.
type UpstreamConfig struct {
	AssetMrn    string
	SpaceMrn    string
	Collector   string
	ApiEndpoint string
	Plugins     []falcon.ClientPlugin
	Incognito   bool
}

// Runtime of all initialized resources
type Runtime struct {
	Registry       *Registry
	cache          *Cache
	Observers      *Observers
	Motor          *motor.Motor
	UpstreamConfig *UpstreamConfig
}

// NewRuntime creates a new runtime from a registry and motor backend
func NewRuntime(registry *Registry, motor *motor.Motor) *Runtime {
	if registry == nil {
		panic("cannot initialize a lumi runtime without a registry")
	}
	if motor == nil {
		panic("cannot initialize a lumi runtime without a motor")
	}

	return &Runtime{
		Registry:  registry,
		Observers: NewObservers(motor),
		Motor:     motor,
		cache:     &Cache{},
	}
}

func args2map(args []interface{}) (*Args, error) {
	if args == nil {
		res := make(Args)
		return &res, nil
	}

	if len(args)%2 == 1 {
		panic("failed to get named argument, it should be supplied as (key, values, ...) and I'm missing a value")
	}

	res := make(Args)
	for i := 0; i < len(args); {
		name, ok := args[i].(string)
		if !ok {
			// TODO: can we get rid of this fmt method?
			return nil, fmt.Errorf("Failed to get named argument, it is not a string field: %#v", args[0])
		}

		res[name] = args[i+1]
		i += 2
	}
	return &res, nil
}

func (ctx *Runtime) createMockResource(name string, cls *ResourceCls) (ResourceType, error) {
	res := MockResource{
		StaticFields: cls.Fields,
		StaticResource: &Resource{
			ResourceID: ResourceID{Id: "", Name: name},
		},
	}
	ctx.Set(res.LumiResource().Name, res.LumiResource().Id, &res)
	return res, nil
}

// CreateResourceWithID creates a new resource instance and force it to have a certain ID
func (ctx *Runtime) CreateResourceWithID(name string, id string, args ...interface{}) (ResourceType, error) {
	r, ok := ctx.Registry.Resources[name]
	if !ok {
		return nil, errors.New("cannot find resource '" + name + "'")
	}

	argsMap, err := args2map(args)
	if err != nil {
		return nil, err
	}

	if r.Factory == nil {
		if len(args) > 0 {
			return nil, errors.New("mock resources don't take any arguments. The resource '" + name + "' doesn't have a resource factory")
		}
		return ctx.createMockResource(name, r)
	}

	// factory not only creates a resource, but may also provide an empty resource
	// with the `Id` field set to look up an existing resource
	res, err := r.Factory(ctx, argsMap)
	if err != nil {
		return nil, errors.New("failed to create resource '" + name + "': " + err.Error())
	}
	if res == nil {
		return nil, errors.New("resource factory produced a nil result for resource '" + name + "'")
	}

	resResource := res.(ResourceType)
	if id == "" {
		id = resResource.LumiResource().Id
	} else {
		resResource.LumiResource().Id = id
	}

	log.Trace().Str("name", name).Str("id", id).Msg("created resource")

	if ex, err := ctx.GetResource(name, id); err == nil {
		resResource = ex
	} else {
		if err := resResource.Validate(); err != nil {
			return nil, errors.New("failed to create resource '" + name + "': " + err.Error())
		}
		ctx.Set(name, id, res)
	}

	return resResource, nil
}

// CreateResource creates a new resource instance taking its name + args
func (ctx *Runtime) CreateResource(name string, args ...interface{}) (ResourceType, error) {
	return ctx.CreateResourceWithID(name, "", args...)
}

// GetRawResource resource instance by name and id
func (ctx *Runtime) getRawResource(name string, id string) (interface{}, bool) {
	res, ok := ctx.cache.Load(name + "\x00" + id)
	if !ok {
		return nil, ok
	}
	return res.Data, ok
}

// GetResource resource instance by name and id
func (ctx *Runtime) GetResource(name string, id string) (ResourceType, error) {
	c, ok := ctx.getRawResource(name, id)
	if !ok {
		return nil, errors.New("cannot find cached resource " + name + " ID: " + id)
	}
	res, ok := c.(ResourceType)
	if !ok {
		return nil, errors.New("cached resource is not of ResourceType for " + name + " ID: " + id)
	}
	return res, nil
}

// Set a resource by name and ID. Must be a valid Resource.
func (ctx *Runtime) Set(name string, id string, resource interface{}) {
	ctx.cache.Store(name+"\x00"+id, &CacheEntry{
		Data:  resource,
		Valid: true,
	})
}

// watch+update => observe it and callback results
// watch+compute => observe it and compute this field when the observed thing changes
// register => build more watch+compute relationships if needed
// trigger => force a field to send a result

// WatchAndUpdate a resource field and call the function if it changes with its current value
func (ctx *Runtime) WatchAndUpdate(r ResourceType, field string, watcherUID string, callback func(res interface{}, err error)) error {
	resource := r.LumiResource()
	// log.Debug().
	// 	Str("src", resource.Name+"\x00"+resource.Id+"\x00"+field).
	// 	Str("watcher", watcherUID).
	// 	Msg("w+u> watch and update")

	// FIXME: calling resource.Fields instead of vv breaks everything!! Make it impossible to do so maybe?
	fieldObj, err := ctx.Registry.Fields(resource.Name)
	if err != nil {
		return errors.New("tried to register field " + field + " in resource " + resource.UID() + ": " + err.Error())
	}
	if fieldObj == nil {
		return errors.New("field object " + field + " in resource " + resource.UID() + " is nil")
	}
	fieldUID := resource.FieldUID(field)

	processResult := func() {
		log.Trace().
			Str("src", resource.Name+"\x00"+resource.Id+"\x00"+field).
			Str("watcher", watcherUID).
			Msg("w+u> process field result")

		data, ok := resource.Cache.Load(field)
		if !ok {
			callback(nil, errors.New("couldn't retrieve value of field \""+field+"\" in resource \""+resource.UID()+"\""))
			return
		}

		callback(data.Data, data.Error)
	}

	isInitial, exists, err := ctx.Observers.Watch(fieldUID, watcherUID, processResult)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// TODO: this is very special handling for when we create a copy of a list
	// resource. in those cases its content (list) has already been filled,
	// but without this block here it will try to compute the entire list from
	// the ground up. It's more of a workaround right now and needs a better
	// solution (eg an indicator for the copied resource?)
	if field == "list" && isInitial {
		data, ok := resource.Cache.Load(field)
		if ok {
			callback(data.Data, data.Error)
		}
	}

	// if the field wasnt registered in the chain of watchers yet,
	// pull all its dependencies in
	if isInitial {
		if err = r.Register(field); err != nil {
			return err
		}

		err = r.Compute(field)
		// normal case most often: we called compute but it depends on something
		// that is not ready
		if _, ok := err.(NotReadyError); ok {
			return nil
		}

		// final case: it is computed and ready to go
		log.Trace().Msg("w+u> initial process result")
		processResult()
		return nil
	}

	data, ok := resource.Cache.Load(field)
	if ok {
		callback(data.Data, data.Error)
	}

	return nil
}

// Unregister will remove all watcher UIDs
func (ctx *Runtime) Unregister(watcherUID string) error {
	log.Trace().Str("watchers", watcherUID).Msg("w+u> unregister")
	return ctx.Observers.UnwatchAll(watcherUID)
}

// WatchAndCompute watches a field in a resource and computes
// another resource + field once once this resource and field has changed
func (ctx *Runtime) WatchAndCompute(src ResourceType, sfield string, dst ResourceType, dfield string) error {
	resource := dst.LumiResource()
	fid := resource.FieldUID(dfield)
	sid := src.LumiResource().FieldUID(sfield)

	isInitial, exists, err := ctx.Observers.Watch(sid, fid, func() {
		// once the source field changes, we recalculate the destination field
		ierr := dst.Compute(dfield)
		// if the field isnt ready, finish this execution
		if _, ok := ierr.(NotReadyError); ok {
			return
		}

		// then we let all the dependent fields know that we just updated this resource field
		ierr = ctx.Trigger(dst, dfield)
		if ierr != nil {
			log.Error().Str("resource+field-uid", fid).Msg("w+c> Failed to trigger resource field: " + ierr.Error())
			return
		}
	})
	log.Trace().
		Str("src", sid).
		Str("dst", fid).
		Bool("initial", isInitial).
		Bool("exists", exists).
		Err(err).
		Msg("w+c> watch and compute")

	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// if the field wasn't registered in the chain of watchers yet,
	// pull all its dependencies in
	if isInitial {
		if err = src.Register(sfield); err != nil {
			log.Error().Err(err).Msg("w+c> initial register failed")
			return err
		}

		err = src.Compute(sfield)
		if err != nil {
			if _, ok := err.(NotReadyError); !ok {
				log.Trace().Err(err).Msg("w+c> initial compute failed")
				return err
			}
		}
	}

	return nil
}

// Trigger a resource-field is a way to request it to calculate its
// value and call the callback. It may use cached values at this point
func (ctx *Runtime) Trigger(r ResourceType, field string) error {
	resource := r.LumiResource()
	if field == "" {
		return errors.New("cannot trigger a resource without specifying a field")
	}

	log.Trace().
		Str("resource", resource.Name+":"+resource.Id).
		Str("field", field).
		Msg("trigger> trigger resource")

	res, ok := resource.Cache.LoadOrStore(field, &CacheEntry{})
	// data in cache means we can go ahead, it's nicely connected already
	// if not it means that the underlying method was never called to compute its value
	// we set the cache to an invalid value to make sure no one else triggers it
	// then we ensure all dependencies send us their results
	if ok {
		entry := res.(*CacheEntry)
		// if it's valid call whatever is listening to this field
		if entry.Valid || entry.Error != nil {
			return ctx.Observers.Trigger(resource.FieldUID(field))
		}
		// if it's not we won't call listening fields yet, because things aren't ready
		return NotReadyError{}
	}

	return r.Compute(field)
}
