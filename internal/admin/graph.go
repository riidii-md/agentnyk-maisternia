package admin

type PhaseNode struct {
	ID        string
	Label     string
	Authority string
}

type PipelineEdge struct {
	From      string
	To        string
	Condition string
	Loop      bool
}

type PipelineGraph struct {
	Nodes  []PhaseNode
	Edges  []PipelineEdge
	GateAt string
}

func DefaultPipeline() PipelineGraph {
	return PipelineGraph{
		Nodes: []PhaseNode{
			{ID: "brief", Label: "BRIEF", Authority: "read_only"},
			{ID: "scout", Label: "SCOUT", Authority: "read_only"},
			{ID: "analyze", Label: "ANALYZE", Authority: "read_only"},
			{ID: "research", Label: "RESEARCH", Authority: "read_only"},
			{ID: "decide", Label: "DECIDE", Authority: "read_only"},
			{ID: "ready", Label: "READY", Authority: "read_only"},
			{ID: "plan", Label: "PLAN", Authority: "read_only"},
			{ID: "prove", Label: "PROVE", Authority: "read_only"},
			{ID: "handoff", Label: "HANDOFF", Authority: "artifact_write"},
			{ID: "run", Label: "RUN", Authority: "workspace_write"},
			{ID: "verify", Label: "VERIFY", Authority: "controlled"},
			{ID: "review", Label: "REVIEW", Authority: "read_only"},
			{ID: "pr", Label: "PR", Authority: "explicit_approval"},
		},
		Edges: []PipelineEdge{
			{From: "brief", To: "scout"},
			{From: "scout", To: "analyze"},
			{From: "analyze", To: "research"},
			{From: "research", To: "decide"},
			{From: "decide", To: "ready"},
			{From: "ready", To: "plan"},
			{From: "plan", To: "prove"},
			{From: "prove", To: "handoff"},
			{From: "handoff", To: "run", Condition: "approval"},
			{From: "run", To: "verify"},
			{From: "verify", To: "review", Condition: "pass"},
			{From: "review", To: "pr", Condition: "pass"},
			{From: "verify", To: "analyze", Condition: "fail", Loop: true},
			{From: "review", To: "run", Condition: "changes", Loop: true},
			{From: "ready", To: "research", Condition: "not ready", Loop: true},
		},
		GateAt: "handoff",
	}
}

func ShapePipeline() PipelineGraph {
	return PipelineGraph{
		Nodes: []PhaseNode{
			{ID: "intake", Label: "INTAKE", Authority: "read_only"},
			{ID: "research", Label: "RESEARCH", Authority: "read_only"},
			{ID: "grill", Label: "GRILL", Authority: "read_only"},
			{ID: "brainstorm", Label: "BRAINSTORM", Authority: "artifact_write"},
			{ID: "challenge", Label: "CHALLENGE", Authority: "read_only"},
			{ID: "decide", Label: "DECIDE", Authority: "artifact_write"},
			{ID: "plan", Label: "PLAN", Authority: "artifact_write"},
			{ID: "final", Label: "FINAL", Authority: "artifact_write"},
		},
		Edges: []PipelineEdge{
			{From: "intake", To: "research"},
			{From: "research", To: "grill"},
			{From: "grill", To: "brainstorm"},
			{From: "brainstorm", To: "challenge"},
			{From: "challenge", To: "decide"},
			{From: "decide", To: "plan"},
			{From: "plan", To: "final", Condition: "human finalization"},
			{From: "grill", To: "research", Condition: "evidence gap", Loop: true},
			{From: "challenge", To: "brainstorm", Condition: "weak options", Loop: true},
			{From: "challenge", To: "grill", Condition: "missing constraint", Loop: true},
		},
		GateAt: "final",
	}
}

func (g PipelineGraph) Node(id string) (PhaseNode, bool) {
	for _, node := range g.Nodes {
		if node.ID == id {
			return node, true
		}
	}
	return PhaseNode{}, false
}
