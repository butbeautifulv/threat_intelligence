export type GraphNode = {
  id: string;
  labels: string[];
  title: string;
  markdown?: string | null;
  kind?: string;
  tags?: string[];
};

export type GraphEdge = {
  id: string;
  source: string;
  target: string;
  type: string;
};

export type GraphData = {
  nodes: GraphNode[];
  edges: GraphEdge[];
};

