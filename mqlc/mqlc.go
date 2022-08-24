package mqlc

import (
	"errors"
	"regexp"
	"sort"
	"strconv"

	vrs "github.com/hashicorp/go-version"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rs/zerolog/log"
	"go.mondoo.com/cnquery"
	"go.mondoo.com/cnquery/llx"
	"go.mondoo.com/cnquery/mqlc/parser"
	v1 "go.mondoo.com/cnquery/mqlc/v1"
	"go.mondoo.com/cnquery/resources"
	"go.mondoo.com/cnquery/resources/packs/all"
	"go.mondoo.com/cnquery/types"
)

type variable struct {
	ref uint64
	typ types.Type
}

type varmap struct {
	blockref uint64
	parent   *varmap
	vars     map[string]variable
}

func newvarmap(blockref uint64, parent *varmap) *varmap {
	return &varmap{
		blockref: blockref,
		parent:   parent,
		vars:     map[string]variable{},
	}
}

func (vm *varmap) lookup(name string) (variable, bool) {
	if v, ok := vm.vars[name]; ok {
		return v, true
	}
	if vm.parent == nil {
		return variable{}, false
	}
	return vm.parent.lookup(name)
}

func (vm *varmap) add(name string, v variable) {
	vm.vars[name] = v
}

func (vm *varmap) len() int {
	return len(vm.vars)
}

type compiler struct {
	Schema    *resources.Schema
	Result    *llx.CodeBundle
	Binding   *variable
	vars      *varmap
	parent    *compiler
	block     *llx.Block
	blockRef  uint64
	blockDeps []uint64
	props     map[string]*llx.Primitive
	comment   string

	// a standalone code is one that doesn't call any of its bindings
	// examples:
	//   file(xyz).content          is standalone
	//   file(xyz).content == _     is not
	standalone bool

	// helps chaining of builtin calls like `if (..) else if (..) else ..`
	prevID string
}

func (c *compiler) isInMyBlock(ref uint64) bool {
	return (ref >> 32) == (c.blockRef >> 32)
}

func (c *compiler) addChunk(chunk *llx.Chunk) {
	c.block.AddChunk(c.Result.CodeV2, c.blockRef, chunk)
}

func (c *compiler) popChunk() (prev *llx.Chunk, isEntrypoint bool, isDatapoint bool) {
	return c.block.PopChunk(c.Result.CodeV2, c.blockRef)
}

func (c *compiler) addArgumentPlaceholder(typ types.Type, checksum string) {
	c.block.AddArgumentPlaceholder(c.Result.CodeV2, c.blockRef, typ, checksum)
}

func (c *compiler) tailRef() uint64 {
	return c.block.TailRef(c.blockRef)
}

// Creates a new block and its accompanying compiler.
// It carries a set of variables that apply within the scope of this block.
func (c *compiler) newBlockCompiler(binding *variable) compiler {
	code := c.Result.CodeV2
	block, ref := code.AddBlock()

	vars := map[string]variable{}
	blockDeps := []uint64{}
	if binding != nil {
		vars["_"] = *binding
		blockDeps = append(blockDeps, binding.ref)
	}

	return compiler{
		Schema:     c.Schema,
		Result:     c.Result,
		Binding:    binding,
		blockDeps:  blockDeps,
		vars:       newvarmap(ref, c.vars),
		parent:     c,
		block:      block,
		blockRef:   ref,
		props:      c.props,
		standalone: true,
	}
}

func addResourceSuggestions(resourceInfos map[string]*resources.ResourceInfo, name string, res *llx.CodeBundle) {
	names := make([]string, len(resourceInfos))
	i := 0
	for key := range resourceInfos {
		names[i] = key
		i++
	}

	suggestedNames := fuzzy.Find(name, names)
	res.Suggestions = make([]*llx.Documentation, len(suggestedNames))
	var info *resources.ResourceInfo
	for i := range suggestedNames {
		field := suggestedNames[i]
		info = resourceInfos[field]
		if info != nil {
			res.Suggestions[i] = &llx.Documentation{
				Field: field,
				Title: info.Title,
				Desc:  info.Desc,
			}
		} else {
			res.Suggestions[i] = &llx.Documentation{
				Field: field,
			}
		}
	}

	sort.SliceStable(res.Suggestions, func(i, j int) bool { return res.Suggestions[i].Field < res.Suggestions[j].Field })
}

func addFieldSuggestions(fields map[string]llx.Documentation, fieldName string, res *llx.CodeBundle) {
	names := make([]string, len(fields))
	i := 0
	for key := range fields {
		names[i] = key
		i++
	}

	suggestedNames := fuzzy.Find(fieldName, names)
	res.Suggestions = make([]*llx.Documentation, len(suggestedNames))
	for i := range suggestedNames {
		info := fields[suggestedNames[i]]
		res.Suggestions[i] = &info
	}

	sort.SliceStable(res.Suggestions, func(i, j int) bool { return res.Suggestions[i].Field < res.Suggestions[j].Field })
}

// func (c *compiler) addAccessor(call *Call, typ types.Type) types.Type {
// 	binding := c.Result.Code.ChunkIndex()
// 	ownerType := c.Result.Code.LastChunk().Type(c.Result.Code)

// 	if call.Accessors != nil {
// 		arg, err := c.compileValue(call.Accessors)
// 		if err != nil {
// 			panic(err.Error())
// 		}

// 		c.Result.Code.AddChunk(&llx.Chunk{
// 			Call: llx.Chunk_FUNCTION,
// 			Id:   "[]",
// 			Function: &llx.Function{
// 				Type:    string(ownerType.Child()),
// 				Binding: binding,
// 				Args:    []*llx.Primitive{arg},
// 			},
// 		})

// 		return ownerType.Child()
// 	}

// 	if call.Params != nil {
// 		panic("We have not yet implemented adding more unnamed function calls")
// 	}

// 	panic("Tried to add accessor calls for a call that has no accessors or params")
// }

// func (c *compiler) addAccessorCalls(calls []*Call, typ types.Type) types.Type {
// 	if calls == nil || len(calls) == 0 {
// 		return typ
// 	}
// 	for i := range calls {
// 		typ = c.addAccessorCall(calls[i], typ)
// 	}
// 	return typ
// }

// compileBlock on a context
func (c *compiler) compileBlock(expressions []*parser.Expression, typ types.Type, bindingRef uint64) (types.Type, error) {
	// For resource, users may indicate to query all fields. It also works for list of resources.
	// This is a special case which is handled here:
	if len(expressions) == 1 && (typ.IsResource() || (typ.IsArray() && typ.Child().IsResource())) {
		x := expressions[0]

		// Special handling for the glob operation on resource fields. It will
		// try to grab all valid fields and return them.
		if x.Operand != nil && x.Operand.Value != nil && x.Operand.Value.Ident != nil && *(x.Operand.Value.Ident) == "*" {
			var fields map[string]llx.Documentation
			if typ.IsArray() {
				fields = availableGlobFields(c, typ.Child())
			} else {
				fields = availableGlobFields(c, typ)
			}

			fieldNames := make([]string, len(fields))
			var i int
			for k := range fields {
				fieldNames[i] = k
				i++
			}
			sort.Strings(fieldNames)

			expressions = []*parser.Expression{}
			for _, v := range fieldNames {
				name := v
				expressions = append(expressions, &parser.Expression{
					Operand: &parser.Operand{
						Value: &parser.Value{Ident: &name},
					},
				})
			}
		}
	}

	fref, blockDeps, _, err := c.blockExpressions(expressions, typ)
	if err != nil {
		return types.Nil, err
	}
	if fref == 0 {
		return typ, nil
	}

	var resultType types.Type
	if typ.IsArray() {
		resultType = types.Array(types.Block)
	} else {
		resultType = types.Block
	}

	args := []*llx.Primitive{llx.FunctionPrimitiveV2(fref)}
	for _, v := range blockDeps {
		if c.isInMyBlock(v) {
			args = append(args, llx.RefPrimitiveV2(v))
		}
	}
	c.blockDeps = append(c.blockDeps, blockDeps...)

	c.addChunk(&llx.Chunk{
		Call: llx.Chunk_FUNCTION,
		Id:   "{}",
		Function: &llx.Function{
			Type:    string(resultType),
			Binding: bindingRef,
			Args:    args,
		},
	})

	return resultType, nil
}

func (c *compiler) compileIfBlock(expressions []*parser.Expression, chunk *llx.Chunk) (types.Type, error) {
	// if `else { .. }` is called, we reset the prevID to indicate there is no
	// more chaining happening
	if c.prevID == "else" {
		c.prevID = ""
	}

	blockCompiler := c.newBlockCompiler(c.Binding)
	err := blockCompiler.compileExpressions(expressions)
	if err != nil {
		return types.Nil, err
	}
	blockCompiler.updateEntrypoints(false)

	block := blockCompiler.block

	// we set this to true, so that we can decide how to handle all following expressions
	if block.SingleValue {
		c.block.SingleValue = true
	}

	// insert a body if we are in standalone mode to return a value
	if len(block.Chunks) == 0 && c.standalone {
		blockCompiler.addChunk(&llx.Chunk{
			Call:      llx.Chunk_PRIMITIVE,
			Primitive: llx.NilPrimitive,
		})
		blockCompiler.addChunk(&llx.Chunk{
			Call: llx.Chunk_FUNCTION,
			Id:   "return",
			Function: &llx.Function{
				Type: string(types.Nil),
				// FIXME: this is gonna crash on c.Binding == nil
				Args: []*llx.Primitive{llx.RefPrimitiveV2(blockCompiler.blockRef | 1)},
			},
		})
		block.SingleValue = true
		block.Entrypoints = []uint64{blockCompiler.blockRef | 2}
	}

	depArgs := []*llx.Primitive{}
	for _, v := range blockCompiler.blockDeps {
		if c.isInMyBlock(v) {
			depArgs = append(depArgs, llx.RefPrimitiveV2(v))
		}
	}

	// the last chunk in this case is the `if` function call
	chunk.Function.Args = append(chunk.Function.Args,
		llx.FunctionPrimitiveV2(blockCompiler.blockRef),
		llx.ArrayPrimitive(depArgs, types.Ref),
	)

	c.blockDeps = append(c.blockDeps, blockCompiler.blockDeps...)

	if len(block.Chunks) != 0 {
		var typeToEnforce types.Type
		if c.block.SingleValue {
			last := block.LastChunk()
			typeToEnforce = last.Type()
		} else {
			typeToEnforce = types.Block
		}

		t, ok := types.Enforce(types.Type(chunk.Function.Type), typeToEnforce)
		if !ok {
			return types.Nil, errors.New("mismatched return type for child block of if-function; make sure all return types are the same")
		}
		chunk.Function.Type = string(t)
	}

	return types.Nil, nil
}

func (c *compiler) compileSwitchCase(expression *parser.Expression, bind *variable, chunk *llx.Chunk) error {
	// for the default case, we get a nil expression
	if expression == nil {
		chunk.Function.Args = append(chunk.Function.Args, llx.BoolPrimitive(true))
		return nil
	}

	prevBind := c.Binding
	c.Binding = bind
	defer func() {
		c.Binding = prevBind
	}()

	argValue, err := c.compileExpression(expression)
	if err != nil {
		return err
	}
	chunk.Function.Args = append(chunk.Function.Args, argValue)
	return nil
}

func (c *compiler) compileSwitchBlock(expressions []*parser.Expression, chunk *llx.Chunk) (types.Type, error) {
	// determine if there is a binding
	// i.e. something inside of those `switch( ?? )` calls
	var bind *variable
	arg := chunk.Function.Args[0]

	// we have to pop the switch chunk from the compiler stack, because it needs
	// to be the last item on the stack. otherwise the last reference (top of stack)
	// will not be pointing to it and an additional entrypoint will be generated

	lastRef := c.block.TailRef(c.blockRef)
	if c.block.LastChunk() != chunk {
		return types.Nil, errors.New("failed to compile switch statement, it wasn't on the top of the compile stack")
	}

	c.block.Chunks = c.block.Chunks[:len(c.block.Chunks)-1]
	c.Result.CodeV2.Checksums[lastRef] = ""

	defer func() {
		c.addChunk(chunk)
	}()

	if types.Type(arg.Type) != types.Unset {
		if types.Type(arg.Type) == types.Ref {
			val, ok := arg.RefV2()
			if !ok {
				return types.Nil, errors.New("could not resolve references of switch argument")
			}
			bind = &variable{
				typ: types.Type(arg.Type),
				ref: val,
			}
		} else {
			c.addChunk(&llx.Chunk{
				Call:      llx.Chunk_PRIMITIVE,
				Primitive: arg,
			})
			ref := c.block.TailRef(c.blockRef)
			bind = &variable{typ: types.Type(arg.Type), ref: ref}
		}
	}

	for i := 0; i < len(expressions); i += 2 {
		err := c.compileSwitchCase(expressions[i], bind, chunk)
		if err != nil {
			return types.Nil, err
		}

		// compile the block of this case/default
		if i+1 >= len(expressions) {
			return types.Nil, errors.New("missing block expression in calling `case`/`default` statement")
		}

		block := expressions[i+1]
		if *block.Operand.Value.Ident != "{}" {
			return types.Nil, errors.New("expected block inside case/default statement")
		}

		expressions := block.Operand.Block

		blockCompiler := c.newBlockCompiler(bind)
		// TODO(jaym): Discuss with dom: don't understand what
		// standalone is used for here
		blockCompiler.standalone = true

		err = blockCompiler.compileExpressions(expressions)
		if err != nil {
			return types.Nil, err
		}
		blockCompiler.updateEntrypoints(false)

		// TODO(jaym): Discuss with dom: v1 seems to hardcore this as
		// single valued
		blockCompiler.block.SingleValue = true

		depArgs := []*llx.Primitive{}
		for _, v := range blockCompiler.blockDeps {
			if c.isInMyBlock(v) {
				depArgs = append(depArgs, llx.RefPrimitiveV2(v))
			}
		}

		chunk.Function.Args = append(chunk.Function.Args,
			llx.FunctionPrimitiveV2(blockCompiler.blockRef),
			llx.ArrayPrimitive(depArgs, types.Ref),
		)

		c.blockDeps = append(c.blockDeps, blockCompiler.blockDeps...)

	}

	// FIXME: I'm pretty sure we don't need this ...
	// c.Result.Code.RefreshChunkChecksum(chunk)

	return types.Nil, nil
}

func (c *compiler) compileUnboundBlock(expressions []*parser.Expression, chunk *llx.Chunk) (types.Type, error) {
	switch chunk.Id {
	case "if":
		t, err := c.compileIfBlock(expressions, chunk)
		if err == nil {
			code := c.Result.CodeV2
			code.Checksums[c.tailRef()] = chunk.ChecksumV2(c.blockRef, code)
		}
		return t, err

	case "switch":
		return c.compileSwitchBlock(expressions, chunk)
	default:
		return types.Nil, errors.New("don't know how to compile unbound block on call `" + chunk.Id + "`")
	}
}

// evaluates the given expressions on a non-array resource
// and creates a function, whose reference is returned
func (c *compiler) blockOnResource(expressions []*parser.Expression, typ types.Type) (uint64, []uint64, bool, error) {
	// binding := &variable{ref: c.tailRef(), typ: typ}
	ogRef := c.tailRef()
	blockCompiler := c.newBlockCompiler(nil)
	blockCompiler.block.AddArgumentPlaceholder(blockCompiler.Result.CodeV2,
		blockCompiler.blockRef, typ, blockCompiler.Result.CodeV2.Checksums[ogRef])
	v := variable{
		ref: blockCompiler.blockRef | 1,
		typ: typ,
	}
	blockCompiler.vars.add("_", v)
	blockCompiler.Binding = &v

	err := blockCompiler.compileExpressions(expressions)
	if err != nil {
		return 0, nil, false, err
	}
	blockCompiler.updateEntrypoints(false)

	return blockCompiler.blockRef, blockCompiler.blockDeps, blockCompiler.standalone, nil
}

// blockExpressions evaluates the given expressions as if called by a block and
// returns the compiled function reference
func (c *compiler) blockExpressions(expressions []*parser.Expression, typ types.Type) (uint64, []uint64, bool, error) {
	if len(expressions) == 0 {
		return 0, nil, false, nil
	}

	if typ.IsArray() {
		return c.blockOnResource(expressions, typ.Child())
	}

	return c.blockOnResource(expressions, typ)
}

// Returns the singular return type of the given block.
// Error if the block has multiple entrypoints (i.e. non singular)
func (c *compiler) blockType(ref uint64) (types.Type, error) {
	block := c.Result.CodeV2.Block(ref)
	if block == nil {
		return types.Nil, errors.New("cannot find block for block ref " + strconv.Itoa(int(ref>>32)))
	}

	if len(block.Entrypoints) != 1 {
		return types.Nil, errors.New("block should only return 1 value (got: " + strconv.Itoa(len(block.Entrypoints)) + ")")
	}

	ep := block.Entrypoints[0]
	chunk := block.Chunks[(ep&0xFFFFFFFF)-1]
	// TODO: this could be a ref! not sure if we can handle that... maybe dereference?
	return chunk.Type(), nil
}

func (c *compiler) dereferenceType(val *llx.Primitive) (types.Type, error) {
	valType := types.Type(val.Type)
	if types.Type(val.Type) != types.Ref {
		return valType, nil
	}

	ref, ok := val.RefV2()
	if !ok {
		return types.Nil, errors.New("found a reference type that doesn't return a reference value")
	}

	chunk := c.Result.CodeV2.Chunk(ref)
	if chunk.Primitive == val {
		return types.Nil, errors.New("recursive reference connections detected")
	}

	if chunk.Primitive != nil {
		return c.dereferenceType(chunk.Primitive)
	}

	valType = chunk.DereferencedTypeV2(c.Result.CodeV2)
	return valType, nil
}

func (c *compiler) unnamedArgs(callerLabel string, init *resources.Init, args []*parser.Arg) ([]*llx.Primitive, error) {
	if len(args) > len(init.Args) {
		return nil, errors.New("Called " + callerLabel +
			" with too many arguments (expected " + strconv.Itoa(len(init.Args)) +
			" but got " + strconv.Itoa(len(args)) + ")")
	}

	// add all calls to the chunk stack
	// collect all their types and call references
	res := make([]*llx.Primitive, len(args)*2)

	for idx := range args {
		arg := args[idx]

		v, err := c.compileExpression(arg.Value)
		if err != nil {
			return nil, errors.New("addResourceCall error: " + err.Error())
		}

		vType := types.Type(v.Type)
		if vType == types.Ref {
			vType, err = c.dereferenceType(v)
			if err != nil {
				return nil, err
			}
		}

		expected := init.Args[idx]
		expectedType := types.Type(expected.Type)
		if vType != expectedType {
			// TODO: We are looking for dict types to see if we can type-cast them
			// This needs massive improvements to dynamically cast them in LLX.
			// For a full description see: https://gitlab.com/mondoolabs/mondoo/-/issues/241
			// This is ONLY a temporary workaround which works in a few cases:
			if vType == types.Dict && expectedType == types.String {
				// we are good, LLX will handle it
			} else {
				return nil, errors.New("Incorrect type on argument " + strconv.Itoa(idx) +
					" in " + callerLabel + ": expected " + expectedType.Label() +
					", got: " + vType.Label())
			}
		}

		res[idx*2] = llx.StringPrimitive(expected.Name)
		res[idx*2+1] = v
	}

	return res, nil
}

func (c *compiler) unnamedResourceArgs(resource *resources.ResourceInfo, args []*parser.Arg) ([]*llx.Primitive, error) {
	if resource.Init == nil {
		return nil, errors.New("cannot find init call for resource " + resource.Id)
	}

	return c.unnamedArgs("resource "+resource.Name, resource.Init, args)
}

// resourceArgs turns the list of arguments for the resource into a list of
// primitives that are used as arguments to initialize that resource
// only works if len(args) > 0 !!
// only works if args are either ALL named or not named !!
func (c *compiler) resourceArgs(resource *resources.ResourceInfo, args []*parser.Arg) ([]*llx.Primitive, error) {
	if args[0].Name == "" {
		return c.unnamedResourceArgs(resource, args)
	}

	res := make([]*llx.Primitive, len(args)*2)
	for idx := range args {
		arg := args[idx]
		field, ok := resource.Fields[arg.Name]
		if !ok {
			return nil, errors.New("resource " + resource.Name + " does not have a field named " + arg.Name)
		}

		v, err := c.compileExpression(arg.Value)
		if err != nil {
			return nil, errors.New("resourceArgs error: " + err.Error())
		}

		vt, err := c.dereferenceType(v)
		if err != nil {
			return nil, err
		}

		ft := types.Type(field.Type)
		if vt != ft {
			return nil, errors.New("Wrong type for field " + arg.Name + " in resource " + resource.Name + ": expected " + ft.Label() + ", got " + vt.Label())
		}

		res[idx*2] = llx.StringPrimitive(arg.Name)
		res[idx*2+1] = v
	}

	return res, nil
}

func (c *compiler) compileBuiltinFunction(h *compileHandler, id string, binding *variable, call *parser.Call) (types.Type, error) {
	if h.compile != nil {
		return h.compile(c, binding.typ, binding.ref, id, call)
	}

	var args []*llx.Primitive

	if call != nil {
		for idx := range call.Function {
			arg := call.Function[idx]
			x, err := c.compileExpression(arg.Value)
			if err != nil {
				return types.Nil, err
			}
			if x != nil {
				args = append(args, x)
			}
		}
	}

	if err := h.signature.Validate(args, c); err != nil {
		return types.Nil, err
	}

	resType := h.typ(binding.typ)
	c.addChunk(&llx.Chunk{
		Call: llx.Chunk_FUNCTION,
		Id:   id,
		Function: &llx.Function{
			Type:    string(resType),
			Binding: binding.ref,
			Args:    args,
		},
	})
	return resType, nil
}

func filterTrailingNullArgs(call *parser.Call) *parser.Call {
	if call == nil {
		return call
	}

	res := parser.Call{
		Comments: call.Comments,
		Ident:    call.Ident,
		Function: call.Function,
		Accessor: call.Accessor,
	}

	args := call.Function
	if len(args) == 0 {
		return &res
	}

	lastIdx := len(args) - 1
	x := args[lastIdx]
	if x.Value.IsEmpty() {
		res.Function = args[0:lastIdx]
	}

	return &res
}

func filterEmptyExpressions(expressions []*parser.Expression) []*parser.Expression {
	res := []*parser.Expression{}
	for i := range expressions {
		exp := expressions[i]
		if exp.IsEmpty() {
			continue
		}
		res = append(res, exp)
	}

	return res
}

// compile a bound identifier to its binding
// example: user { name } , where name is compiled bound to the user
// it will return false if it cannot bind the identifier
func (c *compiler) compileBoundIdentifier(id string, binding *variable, call *parser.Call) (bool, types.Type, error) {
	typ := binding.typ

	if typ.IsResource() {
		resource, ok := c.Schema.Resources[typ.ResourceName()]
		if !ok {
			return true, types.Nil, errors.New("cannot find resource that is called by '" + id + "' of type " + typ.Label())
		}

		fieldinfo, ok := resource.Fields[id]
		if ok {
			if call != nil && len(call.Function) > 0 {
				return true, types.Nil, errors.New("cannot call resource field with arguments yet")
			}

			c.Result.MinMondooVersion = getMinMondooVersion(c.Result.MinMondooVersion, typ.ResourceName(), id)

			// this only happens when we call a field of a bridging resource,
			// in which case we don't call the field (since there is nothing to do)
			// and instead we call the resource directly:
			typ := types.Type(fieldinfo.Type)
			if fieldinfo.IsImplicitResource {
				name := typ.ResourceName()
				c.addChunk(&llx.Chunk{
					Call: llx.Chunk_FUNCTION,
					Id:   name,
				})

				// the new ID is now the full resource call, which is not what the
				// field is originally labeled when we get it, so we have to fix it
				checksum := c.Result.CodeV2.Checksums[c.tailRef()]
				c.Result.Labels.Labels[checksum] = id
				return true, typ, nil
			}

			c.addChunk(&llx.Chunk{
				Call: llx.Chunk_FUNCTION,
				Id:   id,
				Function: &llx.Function{
					Type:    fieldinfo.Type,
					Binding: binding.ref,
				},
			})
			return true, typ, nil
		}
	}

	h, _ := builtinFunction(typ, id)
	if h != nil {
		call = filterTrailingNullArgs(call)
		typ, err := c.compileBuiltinFunction(h, id, binding, call)
		return true, typ, err
	}

	return false, types.Nil, nil
}

// compile a resource from an identifier, trying to find the longest matching resource
// and execute all call functions if there are any
func (c *compiler) compileResource(id string, calls []*parser.Call) (bool, []*parser.Call, types.Type, error) {
	resource, ok := c.Schema.Resources[id]
	if !ok {
		return false, nil, types.Nil, nil
	}

	for len(calls) > 0 && calls[0].Ident != nil {
		nuID := id + "." + (*calls[0].Ident)
		nuResource, ok := c.Schema.Resources[nuID]
		if !ok {
			break
		}
		resource, id = nuResource, nuID
		calls = calls[1:]
	}

	var call *parser.Call
	if len(calls) > 0 && calls[0].Function != nil {
		call = calls[0]
		calls = calls[1:]
	}

	c.Result.MinMondooVersion = getMinMondooVersion(c.Result.MinMondooVersion, id, "")

	typ, err := c.addResource(id, resource, call)
	return true, calls, typ, err
}

func (c *compiler) addResource(id string, resource *resources.ResourceInfo, call *parser.Call) (types.Type, error) {
	var function *llx.Function
	var err error
	typ := types.Resource(id)

	if call != nil && len(call.Function) > 0 {
		function = &llx.Function{Type: string(typ)}
		function.Args, err = c.resourceArgs(resource, call.Function)
		if err != nil {
			return types.Nil, err
		}
	}

	c.addChunk(&llx.Chunk{
		Call:     llx.Chunk_FUNCTION,
		Id:       id,
		Function: function,
	})
	return typ, nil
}

// compileIdentifier within a context of a binding
// 1. global f(): 			expect, ...
// 2. global resource: 	sshd, sshd.config
// 3. bound field: 			user { name }
// x. called field: 		user.name <= not in this scope
func (c *compiler) compileIdentifier(id string, callBinding *variable, calls []*parser.Call) ([]*parser.Call, types.Type, error) {
	var call *parser.Call
	restCalls := calls
	if len(calls) > 0 && calls[0].Function != nil {
		call = calls[0]
		restCalls = calls[1:]
	}

	var typ types.Type
	var err error
	var found bool
	if callBinding != nil {
		// special handling for the `self` operator
		if id == "_" {
			c.standalone = false

			if len(restCalls) == 0 {
				return restCalls, callBinding.typ, nil
			}

			nextCall := restCalls[0]

			if nextCall.Ident != nil {
				calls = restCalls[1:]
				call = nil
				if len(calls) > 0 && calls[0].Function != nil {
					call = calls[0]
				}

				found, typ, err = c.compileBoundIdentifier(*nextCall.Ident, callBinding, call)
				if found {
					if call != nil {
						return restCalls[2:], typ, err
					}
					return restCalls[1:], typ, err
				}
				return nil, types.Nil, errors.New("could not find call _." + (*nextCall.Ident))
			}

			if nextCall.Accessor != nil {
				// turn accessor into a regular function and call that
				fCall := &parser.Call{Function: []*parser.Arg{{Value: nextCall.Accessor}}}
				// accessors are aways builtin functions
				h, _ := builtinFunction(callBinding.typ.Underlying(), "[]")
				if h == nil {
					return nil, types.Nil, errors.New("cannot find '[]' function on type " + callBinding.typ.Label())
				}
				typ, err = c.compileBuiltinFunction(h, "[]", callBinding, fCall)
				if err != nil {
					return nil, types.Nil, err
				}

				if call != nil && len(calls) > 0 {
					calls = calls[1:]
				}

				return restCalls[1:], typ, nil
			}

			return nil, types.Nil, errors.New("not sure how to handle implicit calls around `_`")
		}

		found, typ, err = c.compileBoundIdentifier(id, callBinding, call)
		if found {
			c.standalone = false
			return restCalls, typ, err
		}
	} // end bound functions

	if id == "props" {
		return c.compileProps(call, restCalls, c.Result)
	}

	f := operatorsCompilers[id]
	if f != nil {
		typ, err := f(c, id, call)
		return restCalls, typ, err
	}

	variable, ok := c.vars.lookup(id)
	if ok {
		c.blockDeps = append(c.blockDeps, variable.ref)
		c.addChunk(&llx.Chunk{
			Call:      llx.Chunk_PRIMITIVE,
			Primitive: llx.RefPrimitiveV2(variable.ref),
		})
		return restCalls, variable.typ, nil
	}

	found, restCalls, typ, err = c.compileResource(id, calls)
	if found {
		return restCalls, typ, err
	}

	// suggestions
	if callBinding == nil {
		addResourceSuggestions(c.Schema.Resources, id, c.Result)
		return nil, types.Nil, errors.New("cannot find resource for identifier '" + id + "'")
	}
	addFieldSuggestions(availableFields(c, callBinding.typ), id, c.Result)
	return nil, types.Nil, errors.New("cannot find field or resource '" + id + "' in block for type '" + c.Binding.typ.Label() + "'")
}

// compileProps handles built-in properties for this code
// we will use any properties defined at the compiler-level as type-indicators
func (c *compiler) compileProps(call *parser.Call, calls []*parser.Call, res *llx.CodeBundle) ([]*parser.Call, types.Type, error) {
	if call != nil && len(call.Function) != 0 {
		return nil, types.Nil, errors.New("'props' is not a function")
	}

	if len(calls) == 0 {
		return nil, types.Nil, errors.New("called 'props' without a property, please provide the name you are trying to access")
	}

	nextCall := calls[0]
	restCalls := calls[1:]

	if nextCall.Ident == nil {
		return nil, types.Nil, errors.New("please call 'props' with the name of the property you are trying to access")
	}

	name := *nextCall.Ident
	prim, ok := c.props[name]
	if !ok {
		keys := make(map[string]llx.Documentation, len(c.props))
		for key, prim := range c.props {
			keys[key] = llx.Documentation{
				Field: key,
				Title: key + " (" + types.Type(prim.Type).Label() + ")",
			}
		}

		addFieldSuggestions(keys, name, res)

		return nil, types.Nil, errors.New("cannot find property '" + name + "', please define it first")
	}

	c.addChunk(&llx.Chunk{
		Call: llx.Chunk_PROPERTY,
		Id:   name,
		Primitive: &llx.Primitive{
			Type: prim.Type,
		},
	})

	res.Props[name] = string(prim.Type)

	return restCalls, types.Type(prim.Type), nil
}

// compileValue takes an AST value and compiles it
func (c *compiler) compileValue(val *parser.Value) (*llx.Primitive, error) {
	if val.Bool != nil {
		return llx.BoolPrimitive(bool(*val.Bool)), nil
	}

	if val.Int != nil {
		return llx.IntPrimitive(int64(*val.Int)), nil
	}

	if val.Float != nil {
		return llx.FloatPrimitive(float64(*val.Float)), nil
	}

	if val.String != nil {
		return llx.StringPrimitive(*val.String), nil
	}

	if val.Regex != nil {
		re := string(*val.Regex)
		_, err := regexp.Compile(re)
		if err != nil {
			return nil, errors.New("failed to compile regular expression '" + re + "': " + err.Error())
		}
		return llx.RegexPrimitive(re), nil
	}

	if val.Array != nil {
		arr := make([]*llx.Primitive, len(val.Array))
		var err error
		for i := range val.Array {
			e := val.Array[i]
			arr[i], err = c.compileExpression(e)
			if err != nil {
				return nil, err
			}
		}

		return &llx.Primitive{
			Type:  string(llx.ArrayTypeV2(arr, c.Result.CodeV2)),
			Array: arr,
		}, nil
	}

	if val.Map != nil {
		mapRes := make(map[string]*llx.Primitive, len(val.Map))
		var resType types.Type

		for k, v := range val.Map {
			vv, err := c.compileExpression(v)
			if err != nil {
				return nil, err
			}
			if types.Type(vv.Type) != resType {
				if resType == "" {
					resType = types.Type(vv.Type)
				} else if resType != types.Any {
					resType = types.Any
				}
			}
			mapRes[k] = vv
		}

		if resType == "" {
			resType = types.Unset
		}

		return &llx.Primitive{
			Type: string(types.Map(types.String, resType)),
			Map:  mapRes,
		}, nil
	}

	return llx.NilPrimitive, nil
}

func (c *compiler) compileOperand(operand *parser.Operand) (*llx.Primitive, error) {
	var err error
	var res *llx.Primitive
	var typ types.Type
	var ref uint64

	calls := operand.Calls
	c.comment = operand.Comments

	// value:        bool | string | regex | number | array | map | ident
	// so all simple values are compiled into primitives and identifiers
	// into function calls
	if operand.Value.Ident == nil {
		res, err = c.compileValue(operand.Value)
		if err != nil {
			return nil, err
		}
		typ = types.Type(res.Type)

		if len(calls) > 0 {
			c.addChunk(&llx.Chunk{
				Call: llx.Chunk_PRIMITIVE,
				// no ID for standalone
				Primitive: res,
			})
			ref = c.tailRef()
			res = llx.RefPrimitiveV2(ref)
		}
	} else {
		id := *operand.Value.Ident
		orgcalls := calls
		calls, typ, err = c.compileIdentifier(id, c.Binding, calls)
		if err != nil {
			return nil, err
		}

		ref = c.tailRef()
		if id == "_" && len(orgcalls) == 0 {
			ref = c.Binding.ref
		}

		res = llx.RefPrimitiveV2(ref)
	}

	// operand:      value [ call | accessor | '.' ident ]+ [ block ]
	// dealing with all call types
	for len(calls) > 0 {
		call := calls[0]
		if call.Function != nil {
			return nil, errors.New("don't know how to compile chained functions just yet")
		}

		if call.Comments != "" {
			c.comment = call.Comments
		}

		if call.Accessor != nil {
			// turn accessor into a regular function and call that
			fCall := &parser.Call{Function: []*parser.Arg{{Value: call.Accessor}}}
			// accessors are aways builtin functions
			h, _ := builtinFunction(typ.Underlying(), "[]")
			if h == nil {
				return nil, errors.New("cannot find '[]' function on type " + typ.Label())
			}
			typ, err = c.compileBuiltinFunction(h, "[]", &variable{typ: typ, ref: ref}, fCall)
			if err != nil {
				return nil, err
			}

			if call != nil && len(calls) > 0 {
				calls = calls[1:]
			}
			ref = c.tailRef()
			res = llx.RefPrimitiveV2(ref)
			continue
		}

		if call.Ident != nil {
			var found bool
			var resType types.Type
			id := *call.Ident

			if id == "." {
				// We get this from the parser if the user called the dot-accessor
				// but didn't provide any values at all. It equates a not found and
				// we can now just suggest all fields
				addFieldSuggestions(availableFields(c, typ), "", c.Result)
				return nil, errors.New("missing field name in accessing " + typ.Label())
			}

			calls = calls[1:]
			call = nil
			if len(calls) > 0 && calls[0].Function != nil {
				call = calls[0]
			}

			found, resType, err = c.compileBoundIdentifier(id, &variable{typ: typ, ref: ref}, call)
			if err != nil {
				return nil, err
			}
			if !found {
				addFieldSuggestions(availableFields(c, typ), id, c.Result)
				return nil, errors.New("cannot find field '" + id + "' in " + typ.Label())
			}

			typ = resType
			if call != nil && len(calls) > 0 {
				calls = calls[1:]
			}
			ref = c.tailRef()
			res = llx.RefPrimitiveV2(ref)

			continue
		}

		return nil, errors.New("processed a call without any data")
	}

	if operand.Block != nil {
		// for starters, we need the primitive to exist on the stack,
		// so add it if it's missing
		if x := c.tailRef(); (x & 0xFFFFFFFF) == 0 {
			val, err := c.compileValue(operand.Value)
			if err != nil {
				return nil, err
			}
			c.addChunk(&llx.Chunk{
				Call: llx.Chunk_PRIMITIVE,
				// no ID for standalone
				Primitive: val,
			})
			ref = c.tailRef()
		}

		if typ == types.Nil {
			_, err = c.compileUnboundBlock(operand.Block, c.block.LastChunk())
		} else {
			_, err = c.compileBlock(operand.Block, typ, ref)
		}
		if err != nil {
			return nil, err
		}
		ref = c.tailRef()
		res = llx.RefPrimitiveV2(ref)
	}

	return res, nil
}

func (c *compiler) compileExpression(expression *parser.Expression) (*llx.Primitive, error) {
	if len(expression.Operations) > 0 {
		panic("ran into an expression that wasn't pre-compiled. It has more than 1 value attached to it")
	}
	return c.compileOperand(expression.Operand)
}

func (c *compiler) compileAndAddExpression(expression *parser.Expression) (uint64, error) {
	valc, err := c.compileExpression(expression)
	if err != nil {
		return 0, err
	}

	if types.Type(valc.Type) == types.Ref {
		ref, _ := valc.RefV2()
		return ref, nil
		// nothing to do, the last call was added to the compiled chain
	}

	c.addChunk(&llx.Chunk{
		Call: llx.Chunk_PRIMITIVE,
		// no id for standalone values
		Primitive: valc,
	})

	return c.tailRef(), nil
}

func (c *compiler) compileExpressions(expressions []*parser.Expression) error {
	var err error
	code := c.Result.CodeV2

	// we may have comment-only expressions
	expressions = filterEmptyExpressions(expressions)

	for idx := range expressions {
		if err = expressions[idx].ProcessOperators(); err != nil {
			return err
		}
	}

	var ident string
	var prev string
	for idx := range expressions {
		expression := expressions[idx]
		prev = ident
		ident = ""
		if expression.Operand != nil && expression.Operand.Value != nil && expression.Operand.Value.Ident != nil {
			ident = *expression.Operand.Value.Ident
		}

		if prev == "else" && ident != "if" && c.block.SingleValue {
			// if the previous id is else and its single valued, the following
			// expressions cannot be executed
			return errors.New("single valued block followed by expressions")
		}

		if prev == "if" && ident != "else" && c.block.SingleValue {
			// all following expressions need to be compiled in a block which is
			// conditional to this if-statement unless we're already doing
			// if-else chaining

			c.prevID = "else"
			rest := expressions[idx:]
			_, err := c.compileUnboundBlock(rest, c.block.LastChunk())
			return err
		}

		if ident == "return" {
			// A return statement can only be followed by max 1 more expression
			max := len(expressions)
			if idx+2 < max {
				return errors.New("return statement is followed by too many expressions")
			}

			if idx+1 == max {
				// nothing else coming after this, return nil
			}

			c.block.SingleValue = true
			continue
		}

		// for all other expressions, just compile
		ref, err := c.compileAndAddExpression(expression)
		if err != nil {
			return err
		}

		if prev == "return" {
			prevChunk := code.Chunk(ref)

			c.addChunk(&llx.Chunk{
				Call: llx.Chunk_FUNCTION,
				Id:   "return",
				Function: &llx.Function{
					Type:    string(prevChunk.Type()),
					Binding: 0,
					Args: []*llx.Primitive{
						llx.RefPrimitiveV2(ref),
					},
				},
			})

			c.block.Entrypoints = []uint64{c.block.TailRef(c.blockRef)}
			c.block.SingleValue = true

			return nil
		}

		l := len(c.block.Entrypoints)
		// if the last entrypoint already points to this ref, skip it
		if l != 0 && c.block.Entrypoints[l-1] == ref {
			continue
		}

		c.block.Entrypoints = append(c.block.Entrypoints, ref)

		if code.Checksums[ref] == "" {
			return errors.New("failed to compile expression, ref returned empty checksum ID for ref " + strconv.FormatInt(int64(ref), 10))
		}
	}

	return nil
}

func (c *compiler) updateEntrypoints(collectRefDatapoints bool) {
	// BUG (jaym): collectRefDatapoints prevents us from collecting datapoints.
	// Collecting datapoints for blocks didnt work correctly until 6.7.0.
	// See https://gitlab.com/mondoolabs/mondoo/-/merge_requests/2639
	// We can fix this after some time has passed. If we fix it too soon
	// people will start having their queries fail if a falsy datapoint
	// is collected.

	code := c.Result.CodeV2

	// 1. efficiently remove variable definitions from entrypoints
	varsByRef := make(map[uint64]variable, c.vars.len())
	for name, v := range c.vars.vars {
		if name == "_" {
			// We need to filter this out. It wasn't an assignment declared by the
			// user. We will re-introduce it conceptually once we tackle context
			// information for blocks.
			continue
		}
		varsByRef[v.ref] = v
	}

	max := len(c.block.Entrypoints)
	for i := 0; i < max; {
		ref := c.block.Entrypoints[i]
		if _, ok := varsByRef[ref]; ok {
			c.block.Entrypoints[i], c.block.Entrypoints[max-1] = c.block.Entrypoints[max-1], c.block.Entrypoints[i]
			max--
		} else {
			i++
		}
	}
	if max != len(c.block.Entrypoints) {
		c.block.Entrypoints = c.block.Entrypoints[:max]
	}

	// 2. potentially clean up all inherited entrypoints
	// TODO: unclear if this is necessary because the condition may never be met
	entrypoints := map[uint64]struct{}{}
	for _, ref := range c.block.Entrypoints {
		entrypoints[ref] = struct{}{}
		chunk := code.Chunk(ref)
		if chunk.Function != nil {
			delete(entrypoints, chunk.Function.Binding)
		}
	}

	if !collectRefDatapoints {
		return
	}

	datapoints := map[uint64]struct{}{}
	// 3. resolve operators
	for ref := range entrypoints {
		dps := code.RefDatapoints(ref)
		if dps != nil {
			for i := range dps {
				datapoints[dps[i]] = struct{}{}
			}
		}
	}

	// done
	res := make([]uint64, len(datapoints))
	var idx int
	for ref := range datapoints {
		res[idx] = ref
		idx++
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	c.block.Datapoints = append(c.block.Datapoints, res...)
}

// CompileParsed AST into an executable structure
func (c *compiler) CompileParsed(ast *parser.AST) error {
	err := c.compileExpressions(ast.Expressions)
	if err != nil {
		return err
	}

	c.Result.CodeV2.UpdateID()
	c.updateEntrypoints(true)
	return nil
}

func getMinMondooVersion(current string, resource string, field string) string {
	rd := all.ResourceDocs.Resources[resource]
	var minverDocs string
	if rd != nil {
		minverDocs = rd.MinMondooVersion
		if field != "" {
			f := rd.Fields[field]
			if f != nil && f.MinMondooVersion != "" {
				minverDocs = f.MinMondooVersion
			}
		}
		if current != "" {
			// If the field has a newer version requirement than the current code bundle
			// then update the version requirement to the newest version required.
			docMin, err := vrs.NewVersion(minverDocs)
			curMin, err1 := vrs.NewVersion(current)
			if docMin != nil && err == nil && err1 == nil && docMin.LessThan(curMin) {
				return current
			}
		}
	}
	return minverDocs
}

// CompileAST with a schema into a chunky code
func CompileAST(ast *parser.AST, schema *resources.Schema, props map[string]*llx.Primitive) (*llx.CodeBundle, error) {
	if schema == nil {
		return nil, errors.New("mqlc> please provide a schema to compile this code")
	}

	if props == nil {
		props = map[string]*llx.Primitive{}
	}

	codeBundle := &llx.CodeBundle{
		CodeV2: &llx.CodeV2{
			Checksums: map[uint64]string{},
			// we are initializing it with the first block, which is empty
			Blocks: []*llx.Block{{}},
		},
		Labels: &llx.Labels{
			Labels: map[string]string{},
		},
		Props:            map[string]string{},
		Version:          mondoo.ApiVersion(),
		MinMondooVersion: "",
	}

	c := compiler{
		Schema:     schema,
		Result:     codeBundle,
		vars:       newvarmap(1<<32, nil),
		parent:     nil,
		blockRef:   1 << 32,
		block:      codeBundle.CodeV2.Blocks[0],
		props:      props,
		standalone: true,
	}

	return c.Result, c.CompileParsed(ast)
}

// Compile a code piece against a schema into chunky code
func compile(input string, schema *resources.Schema, props map[string]*llx.Primitive) (*llx.CodeBundle, error) {
	// remove leading whitespace; we are re-using this later on
	input = Dedent(input)

	ast, err := parser.Parse(input)
	if ast == nil {
		return nil, err
	}

	// Special handling for parser errors: We still try to compile it because
	// we want to get any compiler suggestions for auto-complete / fixing it.
	// That said, we must return an error either way.
	if err != nil {
		res, _ := CompileAST(ast, schema, props)
		return res, err
	}

	res, err := CompileAST(ast, schema, props)
	if err != nil {
		return res, err
	}

	err = UpdateLabels(res.CodeV2, res.Labels, schema)
	if err != nil {
		return res, err
	}
	if len(res.Labels.Labels) == 0 {
		res.Labels.Labels = nil
	}

	err = UpdateAssertions(res)
	if err != nil {
		return res, err
	}

	res.Source = input
	return res, nil
}

func Compile(input string, schema *resources.Schema, features cnquery.Features, props map[string]*llx.Primitive) (*llx.CodeBundle, error) {
	if features.IsActive(cnquery.PiperCode) {
		res, err := compile(input, schema, props)
		if err != nil {
			return res, err
		}

		if res.CodeV2 == nil || res.CodeV2.Id == "" {
			return res, errors.New("failed to compile: received an unspecified empty code structure")
		}

		return res, nil
	}

	res, err := v1.Compile(input, schema, props)
	if err != nil {
		return res, err
	}

	if res.DeprecatedV5Code == nil || res.DeprecatedV5Code.Id == "" {
		return res, errors.New("failed to compile: received an unspecified empty code structure (v1)")
	}

	return res, nil
}

// MustCompile a code piece that should not fail (otherwise panic)
func MustCompile(input string, schema *resources.Schema, features cnquery.Features, props map[string]*llx.Primitive) *llx.CodeBundle {
	res, err := Compile(input, schema, features, props)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to compile")
	}
	return res
}
