package main

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"sort"
	"strings"
)
import "github.com/metoro-io/mcp-golang"
import "github.com/metoro-io/mcp-golang/transport/stdio"

// MCPServer represents the Model Context Protocol server config

type ConvertTo struct {
	ID                string
	DestinationFormat string
}

type ConvertFrom struct {
	ID                string
	DestinationFormat string
}

func RunMCP() error {
	done := make(chan struct{})
	delete(decodings, "binary")
	delete(encodings, "binary")
	delete(encodings, "remote-id")
	server := mcp_golang.NewServer(stdio.NewStdioServerTransport())

	satellites := map[string]string{
		"us1":      "12EayRS2V1kEsWESU9QMRseFhdxYxKicsiFmxrsLZHeLUtdps3S",
		"eu1":      "12L9ZFwhzVpuEKMUNUqkaTLGzwY9G24tbiigLiXpmZWKwmcNDDs",
		"ap1":      "121RTSDpyNZVcEU84Ticf2L1ntiuUimbWgfATz21tuvgk3vzoA6",
		"slc":      "1wFTAgs9DP5RSnCqKV1eLf6N9wtk4EAtmN5DpSxcs8EjT69tGE",
		"saltlake": "1wFTAgs9DP5RSnCqKV1eLf6N9wtk4EAtmN5DpSxcs8EjT69tGE",
	}
	for name, id := range satellites {
		err := server.RegisterResource("storj-id://"+name, strings.ToUpper(name)+" Satellite ID", "NodeID of Storj "+strings.ToUpper(name)+" Satellite", "", func() (*mcp_golang.ResourceResponse, error) {
			return &mcp_golang.ResourceResponse{
				Contents: []*mcp_golang.EmbeddedResource{
					mcp_golang.NewTextResourceContent("storj-id://"+name, id, "").EmbeddedResource,
				},
			}, nil
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	err := server.RegisterTool("convert_to", "Convert one ID of Storj world from one representation to other (such as NodeID, HEX, Base64, Base58, Base32 or PieceID)", func(input ConvertTo) (*mcp_golang.ToolResponse, error) {
		input.DestinationFormat = strings.ToLower(input.DestinationFormat)
		var results []*mcp_golang.Content

		encoding, found := encodings[input.DestinationFormat]
		if !found {
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Unknown encoding format %s. Use one of %s.",
				input.DestinationFormat, strings.Join(maps.Keys(encodings), ",")))), nil
		}

		keys := maps.Keys(decodings)
		sort.Strings(keys)

		used := map[string]bool{}
		for _, id := range keys {
			if id == input.DestinationFormat {
				continue
			}
			decoded, err := decodings[id](input.ID)
			if err == nil {
				decoded := encoding(decoded)
				if decoded != "" {
					if _, found := used[decoded]; found {
						continue
					}
					used[decoded] = true
					results = append(results, mcp_golang.NewTextContent(fmt.Sprintf("ID %s is converted from %s to %s: %s", input.ID, id, input.DestinationFormat, decoded)))
				}
			}
		}

		if len(results) == 0 {
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Couldn't read the source string with any of the known encodings: %s", strings.Join(keys, ",")))), nil
		}
		return mcp_golang.NewToolResponse(results...), nil
	})
	if err != nil {
		return errors.WithStack(err)
	}

	err = server.Serve()
	if err != nil {
		return errors.WithStack(err)
	}

	<-done
	return nil
}
