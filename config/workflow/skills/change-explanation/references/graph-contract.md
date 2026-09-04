# PR Lens graph contract for change explanations

This reference covers the small subset needed by `/work-explain-change`. The
installed PR Lens validator is authoritative; unknown fields are rejected.

Use schema `0.1.0` with `kind: graph`. Each node belongs to a declared lane.
Every edge endpoint, flow participant, and flow-message endpoint names a
declared node. IDs are unique readable identifiers. File paths are
repository-relative POSIX paths without `..` segments.

Allowed deltas are `added`, `modified`, `removed`, and `unchanged`. Useful node
kinds include `service`, `app`, `module`, `function`, `route`, `job`, `queue`,
`datastore`, `cache`, `external`, `ui`, `config`, `test`, `package`, and
`other`. Architecture edge kinds include `call`, `http`, `rpc`, `event`,
`queue`, `data`, `dependency`, `render`, and `other`. Flow message kinds are
`sync`, `async`, `return`, and `self`; only `self` may use the same node for
`from` and `to`.

A minimal architecture and data-flow document looks like this:

```json
{
  "schemaVersion": "0.1.0",
  "kind": "graph",
  "title": "Authenticated local change diagrams",
  "summary": "The document reader resolves a registered SVG through an authenticated same-origin media route.",
  "lenses": ["architecture", "data-flow"],
  "provenance": {
    "repo": {"owner": "example", "name": "desk", "host": "github.com"},
    "base": {"sha": "1111111", "ref": "main"},
    "head": {"sha": "2222222", "ref": "feature/local-media"},
    "generator": {"name": "work-explain-change", "version": "0.1.0"}
  },
  "lanes": [
    {"id": "document", "label": "Document"},
    {"id": "server", "label": "Desk server"},
    {"id": "browser", "label": "Browser"}
  ],
  "nodes": [
    {
      "id": "markdown-image",
      "label": "Markdown image",
      "kind": "other",
      "delta": "unchanged",
      "lane": "document",
      "files": [{"path": "reports/explanation.md"}],
      "badges": []
    },
    {
      "id": "media-route",
      "label": "Authenticated media route",
      "kind": "route",
      "delta": "added",
      "lane": "server",
      "files": [{"path": "src/server.ts"}],
      "badges": ["same origin"]
    },
    {
      "id": "reader",
      "label": "Document reader",
      "kind": "ui",
      "delta": "modified",
      "lane": "browser",
      "files": [{"path": "src/web-client.ts"}],
      "badges": []
    }
  ],
  "edges": [
    {
      "id": "image-to-route",
      "from": "markdown-image",
      "to": "media-route",
      "kind": "render",
      "delta": "added",
      "label": "rewrite local src",
      "emphasis": "hero",
      "animated": true,
      "files": [{"path": "src/server.ts"}]
    },
    {
      "id": "route-to-reader",
      "from": "media-route",
      "to": "reader",
      "kind": "http",
      "delta": "added",
      "label": "SVG response",
      "emphasis": "normal",
      "animated": true,
      "files": [{"path": "src/server.ts"}]
    }
  ],
  "flows": [
    {
      "id": "load-diagram",
      "title": "Load the local diagram",
      "delta": "added",
      "participants": [
        {"node": "reader", "label": "reader"},
        {"node": "media-route", "label": "media route"}
      ],
      "messages": [
        {
          "id": "request-svg",
          "from": "reader",
          "to": "media-route",
          "label": "GET registered asset",
          "kind": "sync",
          "delta": "added",
          "animated": true,
          "files": [{"path": "src/server.ts"}]
        },
        {
          "id": "return-svg",
          "from": "media-route",
          "to": "reader",
          "label": "sandboxed SVG",
          "kind": "return",
          "delta": "added",
          "animated": true,
          "files": [{"path": "src/server.ts"}]
        }
      ]
    }
  ],
  "views": [
    {
      "id": "architecture-overview",
      "title": "Architecture overview",
      "lens": "architecture",
      "scope": {"kind": "all"},
      "defaultOpen": true,
      "children": []
    },
    {
      "id": "load-diagram-flow",
      "title": "Diagram loading flow",
      "lens": "data-flow",
      "scope": {
        "kind": "selection",
        "lanes": [],
        "nodes": ["reader", "media-route"],
        "edges": [],
        "flows": ["load-diagram"]
      },
      "defaultOpen": true,
      "children": []
    }
  ],
  "layout": {
    "direction": "right",
    "laneOrder": ["document", "server", "browser"]
  }
}
```

Omit the `data-flow` lens, `flows`, and data-flow view when the change has no
meaningful ordered interaction. Use `animated: true` only on the few changed
connections or flow messages whose movement aids understanding. Include
unchanged neighbours when they explain blast radius, but do not turn the graph
into a symbol inventory.
