package resources

import (
	"fmt"

	"go.mondoo.io/mondoo/lumi/resources/windows"
)

func (p *lumiAuditpol) id() (string, error) {
	return "auditpol", nil
}

func (p *lumiAuditpol) GetList() ([]interface{}, error) {
	cmd, err := p.MotorRuntime.Motor.Transport.RunCommand("auditpol /get /category:* /r")
	if err != nil {
		return nil, fmt.Errorf("could not run auditpol")
	}

	entries, err := windows.ParseAuditpol(cmd.Stdout)
	if err != nil {
		return nil, err
	}

	auditPolEntries := make([]interface{}, len(entries))
	for i := range entries {
		entry := entries[i]
		lumiAuditpolEntry, err := p.MotorRuntime.CreateResource("auditpol.entry",
			"machinename", entry.MachineName,
			"policytarget", entry.PolicyTarget,
			"subcategory", entry.Subcategory,
			"subcategoryguid", entry.SubcategoryGUID,
			"inclusionsetting", entry.InclusionSetting,
			"exclusionsetting", entry.ExclusionSetting,
		)
		if err != nil {
			return nil, err
		}
		auditPolEntries[i] = lumiAuditpolEntry.(AuditpolEntry)
	}

	return auditPolEntries, nil
}

func (p *lumiAuditpolEntry) id() (string, error) {
	return p.Subcategoryguid()
}
