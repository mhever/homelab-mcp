11d10
< 	"github.com/mhever/homelab-mcp/internal/mcputil"
17,18d15
< type VersionArgs struct{}
< 
30,37d26
< 
< 	// Version tool
< 	mcp.AddTool(server, &mcp.Tool{
< 		Name:        "version",
< 		Description: "Returns the server version",
< 	}, func(ctx context.Context, req *mcp.CallToolRequest, args VersionArgs) (*mcp.CallToolResult, any, error) {
< 		return mcputil.TextResult("1.0.0")
< 	})
