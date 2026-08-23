package main

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
	"github.com/claytonharbour/proseforge-workbench/internal/bundle"
	"github.com/claytonharbour/proseforge-workbench/internal/story"
)

func registerBundleTools(s *server.MCPServer, r *clientResolver) {
	// bundle_list — List bundles
	s.AddTool(
		tool("bundle_list",
			mcp.WithDescription(
				"List the authenticated user's bundles. Returns bundle ID, name, intro, "+
					"entry count, and timestamps."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := bundle.NewService(client, bundle.WithLogger(r.logger))
			result, err := svc.List(ctx)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// bundle_get — Get bundle with entries
	s.AddTool(
		tool("bundle_get",
			mcp.WithDescription(
				"Get a bundle with its entries. Returns bundle metadata, cover image, and "+
					"ordered list of entries (story ID, title, comment, book number, images)."),
			mcp.WithString("bundle_id", mcp.Required(), mcp.Description("Bundle ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			bundleID, err := requireArg(req, "bundle_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := bundle.NewService(client, bundle.WithLogger(r.logger))
			result, err := svc.Get(ctx, bundleID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// bundle_create — Create a bundle
	s.AddTool(
		tool("bundle_create",
			mcp.WithDescription(
				"Create a new bundle. Bundles are packaging recipes — add stories as entries, "+
					"arrange in order, add interstitial text between entries, and export as "+
					"EPUB/PDF/markdown/JSON. For the full workflow, read docs://bundle-workflow."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Bundle name")),
			mcp.WithString("intro", mcp.Description("Introduction text for the bundle (markdown)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			name, err := requireArg(req, "name")
			if err != nil {
				return toolError(err), nil
			}
			svc := bundle.NewService(client, bundle.WithLogger(r.logger))
			result, err := svc.Create(ctx, name, optionalArg(req, "intro"))
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// bundle_update — Update a bundle
	s.AddTool(
		tool("bundle_update",
			mcp.WithDescription("Update a bundle's name or introduction. Only provided fields are changed."),
			mcp.WithString("bundle_id", mcp.Required(), mcp.Description("Bundle ID")),
			mcp.WithString("name", mcp.Description("New bundle name")),
			mcp.WithString("intro", mcp.Description("New introduction text (markdown)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			bundleID, err := requireArg(req, "bundle_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := bundle.NewService(client, bundle.WithLogger(r.logger))
			if err := svc.Update(ctx, bundleID, optionalArg(req, "name"), optionalArg(req, "intro")); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Bundle updated."), nil
		},
	)

	// bundle_delete — Delete a bundle
	s.AddTool(
		tool("bundle_delete",
			mcp.WithDescription(
				"Delete a bundle. This removes the packaging recipe only — the stories "+
					"themselves are not affected."),
			mcp.WithString("bundle_id", mcp.Required(), mcp.Description("Bundle ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{DestructiveHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			bundleID, err := requireArg(req, "bundle_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := bundle.NewService(client, bundle.WithLogger(r.logger))
			if err := svc.Delete(ctx, bundleID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Bundle deleted."), nil
		},
	)

	// bundle_entry_add — Add a story to a bundle
	s.AddTool(
		tool("bundle_entry_add",
			mcp.WithDescription(
				"Add a story to a bundle as an entry. Optionally include a comment — "+
					"interstitial text shown with the story in the exported bundle "+
					"(forewords, interludes, transitions)."),
			mcp.WithString("bundle_id", mcp.Required(), mcp.Description("Bundle ID")),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID to add")),
			mcp.WithString("comment", mcp.Description("Interstitial transition text for this entry (markdown)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			bundleID, err := requireArg(req, "bundle_id")
			if err != nil {
				return toolError(err), nil
			}
			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := bundle.NewService(client, bundle.WithLogger(r.logger))
			result, err := svc.AddEntry(ctx, bundleID, storyID, optionalArg(req, "comment"))
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// bundle_entry_update — Update entry transition text
	s.AddTool(
		tool("bundle_entry_update",
			mcp.WithDescription("Update the transition text for a bundle entry. Called 'comment' here, 'Transition' in the UI — same field. This text appears with the entry's story in the export."),
			mcp.WithString("bundle_id", mcp.Required(), mcp.Description("Bundle ID")),
			mcp.WithString("entry_id", mcp.Required(), mcp.Description("Entry ID (from bundle_get)")),
			mcp.WithString("comment", mcp.Required(), mcp.Description("New transition text (markdown)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			bundleID, err := requireArg(req, "bundle_id")
			if err != nil {
				return toolError(err), nil
			}
			entryID, err := requireArg(req, "entry_id")
			if err != nil {
				return toolError(err), nil
			}
			comment, err := requireArg(req, "comment")
			if err != nil {
				return toolError(err), nil
			}
			svc := bundle.NewService(client, bundle.WithLogger(r.logger))
			if err := svc.UpdateEntry(ctx, bundleID, entryID, comment); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Bundle entry updated."), nil
		},
	)

	// bundle_entry_remove — Remove an entry
	s.AddTool(
		tool("bundle_entry_remove",
			mcp.WithDescription(
				"Remove an entry from a bundle. The story itself is not affected."),
			mcp.WithString("bundle_id", mcp.Required(), mcp.Description("Bundle ID")),
			mcp.WithString("entry_id", mcp.Required(), mcp.Description("Entry ID (from bundle_get)")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{DestructiveHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			bundleID, err := requireArg(req, "bundle_id")
			if err != nil {
				return toolError(err), nil
			}
			entryID, err := requireArg(req, "entry_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := bundle.NewService(client, bundle.WithLogger(r.logger))
			if err := svc.RemoveEntry(ctx, bundleID, entryID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Bundle entry removed."), nil
		},
	)

	// bundle_entry_reorder — Reorder entries
	s.AddTool(
		tool("bundle_entry_reorder",
			mcp.WithDescription(
				"Set the display order of entries in a bundle. Pass all entry IDs in the "+
					"desired order — first becomes entry 1, second becomes entry 2, etc."),
			mcp.WithString("bundle_id", mcp.Required(), mcp.Description("Bundle ID")),
			mcp.WithString("entry_ids", mcp.Required(), mcp.Description("Ordered JSON array of entry IDs")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			bundleID, err := requireArg(req, "bundle_id")
			if err != nil {
				return toolError(err), nil
			}
			idsStr, err := requireArg(req, "entry_ids")
			if err != nil {
				return toolError(err), nil
			}

			var ids []string
			if err := json.Unmarshal([]byte(idsStr), &ids); err != nil {
				return mcp.NewToolResultError("entry_ids must be a JSON array of strings"), nil
			}

			svc := bundle.NewService(client, bundle.WithLogger(r.logger))
			if err := svc.ReorderEntries(ctx, bundleID, ids); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Bundle entry order updated."), nil
		},
	)

	// bundle_export — Export a bundle
	s.AddTool(
		tool("bundle_export",
			mcp.WithDescription(
				"Export a bundle in the specified format. Assembles all entries with "+
					"interstitial comments and images into a single artifact."),
			mcp.WithString("bundle_id", mcp.Required(), mcp.Description("Bundle ID")),
			mcp.WithString("format", mcp.Required(), mcp.Description("Export format: epub, json, markdown, pdf")),
			mcp.WithString("out_path", mcp.Description("Where to write the file. Required for epub and pdf; ignored for json and markdown. Supports ~/.")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			bundleID, err := requireArg(req, "bundle_id")
			if err != nil {
				return toolError(err), nil
			}
			format, err := requireArg(req, "format")
			if err != nil {
				return toolError(err), nil
			}

			switch format {
			case "epub", "json", "markdown", "pdf":
				// valid
			default:
				return mcp.NewToolResultError("format must be one of: epub, json, markdown, pdf"), nil
			}

			// Same defect as story_export (#265): epub and pdf are binary and
			// cannot be returned inline. Refuse early rather than render first.
			if isBinaryExportFormat(format) && optionalArg(req, "out_path") == "" {
				return binaryExportResult(format, "", nil)
			}

			svc := bundle.NewService(client, bundle.WithLogger(r.logger))
			result, err := svc.Export(ctx, bundleID, format)
			if err != nil {
				return toolError(err, client), nil
			}
			if isBinaryExportFormat(format) {
				return binaryExportResult(format, optionalArg(req, "out_path"), result)
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// bundle_image_generate — Generate an image and set as bundle cover
	s.AddTool(
		tool("bundle_image_generate",
			mcp.WithDescription(
				"Generate an AI image and set it as the bundle cover. Pass a prompt "+
					"describing the cover image. Costs 2 credits. Generation is async — "+
					"the cover will appear on the bundle when ready."),
			mcp.WithString("bundle_id", mcp.Required(), mcp.Description("Bundle ID")),
			mcp.WithString("prompt", mcp.Required(), mcp.Description("Image prompt describing the cover")),
			mcp.WithNumber("width", mcp.Description("Image width in pixels")),
			mcp.WithNumber("height", mcp.Description("Image height in pixels")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}

			bundleID, err := requireArg(req, "bundle_id")
			if err != nil {
				return toolError(err), nil
			}
			prompt, err := requireArg(req, "prompt")
			if err != nil {
				return toolError(err), nil
			}

			body := gen.HandlersGenerateImageRequest{
				UserPrompt: &prompt,
			}
			if w := optionalIntArg(req, "width", 0); w > 0 {
				body.Width = &w
			}
			if h := optionalIntArg(req, "height", 0); h > 0 {
				body.Height = &h
			}

			imgSvc := story.NewService(client, story.WithLogger(r.logger))
			result, err := imgSvc.GenerateImage(ctx, body)
			if err != nil {
				return toolError(err, client), nil
			}

			// Parse image ID and set as cover
			var generated struct {
				Image struct {
					ID string `json:"id"`
				} `json:"image"`
			}
			if err := json.Unmarshal(result, &generated); err == nil && generated.Image.ID != "" {
				bundleSvc := bundle.NewService(client, bundle.WithLogger(r.logger))
				if err := bundleSvc.SetCover(ctx, bundleID, generated.Image.ID); err != nil {
					return toolError(err, client), nil
				}
			}

			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// bundle_image_upload — Upload an image and set as bundle cover
	s.AddTool(
		tool("bundle_image_upload",
			mcp.WithDescription(
				"Upload a pre-made image from disk and set it as the bundle cover. "+
					"Max 10 MiB (10,485,760 bytes — binary, not 10,000,000), jpeg/png/webp."),
			mcp.WithString("bundle_id", mcp.Required(), mcp.Description("Bundle ID")),
			mcp.WithString("file_path", mcp.Required(), mcp.Description("Absolute path to the image file on disk")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}

			bundleID, err := requireArg(req, "bundle_id")
			if err != nil {
				return toolError(err), nil
			}
			filePath, err := requireArg(req, "file_path")
			if err != nil {
				return toolError(err), nil
			}

			contentType, body, err := buildMultipartUpload(filePath)
			if err != nil {
				return toolError(err), nil
			}

			imgSvc := story.NewService(client, story.WithLogger(r.logger))
			result, err := imgSvc.UploadImage(ctx, contentType, body)
			if err != nil {
				return toolError(err, client), nil
			}

			// Parse image ID and set as cover
			var uploaded struct {
				Image struct {
					ID string `json:"id"`
				} `json:"image"`
			}
			if err := json.Unmarshal(result, &uploaded); err != nil {
				return toolError(err, client), nil
			}
			if uploaded.Image.ID == "" {
				return mcp.NewToolResultError("Upload succeeded but response contained no image ID"), nil
			}

			bundleSvc := bundle.NewService(client, bundle.WithLogger(r.logger))
			if err := bundleSvc.SetCover(ctx, bundleID, uploaded.Image.ID); err != nil {
				return toolError(err, client), nil
			}

			return jsonResult(uploaded)
		},
	)
}
