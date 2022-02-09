package elasticsearch

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// ListCommand calls the Fastly API to list Elasticsearch logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListElasticsearchInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("list", "List Elasticsearch endpoints on a Fastly service version")
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	elasticsearchs, err := c.Globals.APIClient.ListElasticsearch(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		if c.json {
			data, err := json.Marshal(elasticsearchs)
			if err != nil {
				return err
			}
			fmt.Fprint(out, string(data))
			return nil
		}

		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, elasticsearch := range elasticsearchs {
			tw.AddLine(elasticsearch.ServiceID, elasticsearch.ServiceVersion, elasticsearch.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, elasticsearch := range elasticsearchs {
		fmt.Fprintf(out, "\tElasticsearch %d/%d\n", i+1, len(elasticsearchs))
		fmt.Fprintf(out, "\t\tService ID: %s\n", elasticsearch.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", elasticsearch.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", elasticsearch.Name)
		fmt.Fprintf(out, "\t\tIndex: %s\n", elasticsearch.Index)
		fmt.Fprintf(out, "\t\tURL: %s\n", elasticsearch.URL)
		fmt.Fprintf(out, "\t\tPipeline: %s\n", elasticsearch.Pipeline)
		fmt.Fprintf(out, "\t\tTLS CA certificate: %s\n", elasticsearch.TLSCACert)
		fmt.Fprintf(out, "\t\tTLS client certificate: %s\n", elasticsearch.TLSClientCert)
		fmt.Fprintf(out, "\t\tTLS client key: %s\n", elasticsearch.TLSClientKey)
		fmt.Fprintf(out, "\t\tTLS hostname: %s\n", elasticsearch.TLSHostname)
		fmt.Fprintf(out, "\t\tUser: %s\n", elasticsearch.User)
		fmt.Fprintf(out, "\t\tPassword: %s\n", elasticsearch.Password)
		fmt.Fprintf(out, "\t\tFormat: %s\n", elasticsearch.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", elasticsearch.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", elasticsearch.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", elasticsearch.Placement)

	}
	fmt.Fprintln(out)

	return nil
}
