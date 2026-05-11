'use client';

import dynamic from 'next/dynamic';
import { useEffect, useMemo, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import type { GraphData, GraphNode } from './types';

const ForceGraph2D = dynamic(() => import('react-force-graph-2d'), { ssr: false });

function nodeTitle(n: GraphNode) {
  return n.title || n.id;
}

export default function GraphExplorer() {
  const [graph, setGraph] = useState<GraphData>({ nodes: [], edges: [] });
  const [loading, setLoading] = useState(true);
  const [q, setQ] = useState('');
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [selected, setSelected] = useState<GraphNode | null>(null);
  const [selectedMarkdown, setSelectedMarkdown] = useState<string>('');

  useEffect(() => {
    let cancelled = false;
    async function load() {
      setLoading(true);
      try {
        const res = await fetch('/api/graph');
        const data = (await res.json()) as GraphData;
        if (!cancelled) setGraph(data);
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    let cancelled = false;
    async function loadNode(id: string) {
      const res = await fetch(`/api/node/${encodeURIComponent(id)}`);
      const data = (await res.json()) as { node: GraphNode; markdown: string };
      if (cancelled) return;
      setSelected(data.node);
      setSelectedMarkdown(data.markdown || '');
    }
    if (selectedId) void loadNode(selectedId);
    return () => {
      cancelled = true;
    };
  }, [selectedId]);

  const filteredNodes = useMemo(() => {
    const s = q.trim().toLowerCase();
    if (!s) return graph.nodes.slice(0, 200);
    return graph.nodes
      .filter((n) => `${n.title} ${n.id} ${n.labels.join(' ')}`.toLowerCase().includes(s))
      .slice(0, 200);
  }, [graph.nodes, q]);

  const graphForViz = useMemo(() => {
    const nodeSet = new Set(filteredNodes.map((n) => n.id));
    const edges = graph.edges.filter((e) => nodeSet.has(e.source) && nodeSet.has(e.target));
    return { nodes: filteredNodes, links: edges.map((e) => ({ ...e, source: e.source, target: e.target })) };
  }, [filteredNodes, graph.edges]);

  return (
    <div className="app">
      <section className="card">
        <div className="cardHeader">
          <div className="title">Nodes</div>
          <span className="pill">{loading ? 'loading…' : `${graph.nodes.length} total`}</span>
        </div>
        <div className="content">
          <input
            className="input"
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Search (label/title/id)…"
          />
          <div style={{ height: 10 }} />
          <div className="list">
            {filteredNodes.map((n) => (
              <div
                key={n.id}
                className={`listItem ${selectedId === n.id ? 'listItemActive' : ''}`}
                onClick={() => setSelectedId(n.id)}
                role="button"
                tabIndex={0}
              >
                <div className="name">{nodeTitle(n)}</div>
                <div className="meta">
                  <span className="pill">{n.labels[0] ?? 'Node'}</span>
                  <span className="pill">{n.id}</span>
                </div>
              </div>
            ))}
          </div>
          {!filteredNodes.length && <div className="hint">No matches.</div>}
        </div>
      </section>

      <section className="card">
        <div className="cardHeader">
          <div className="title">Graph</div>
          <span className="pill">{filteredNodes.length} shown</span>
        </div>
        <div style={{ height: 'calc(100% - 49px)' }}>
          <ForceGraph2D
            graphData={graphForViz as any}
            backgroundColor="rgba(0,0,0,0)"
            nodeLabel={(n: any) => nodeTitle(n as GraphNode)}
            nodeAutoColorBy={(n: any) => (n.labels?.[0] ?? 'Node')}
            linkColor={() => 'rgba(255,255,255,.20)'}
            linkWidth={1}
            onNodeClick={(n: any) => setSelectedId((n as GraphNode).id)}
          />
        </div>
      </section>

      <section className="card right">
        <div className="cardHeader">
          <div className="title">Markdown</div>
          {selected ? <span className="pill">{selected.labels[0] ?? 'Node'}</span> : <span className="pill">none</span>}
        </div>
        <div className="content">
          {!selectedId && <div className="hint">Click a node to render its markdown (or fall back to properties).</div>}
          {selectedId && !selected && <div className="hint">Loading…</div>}
          {selected && (
            <>
              <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginBottom: 12 }}>
                <span className="pill">{selected.id}</span>
                {selected.labels.map((l) => (
                  <span key={l} className="pill">
                    {l}
                  </span>
                ))}
              </div>
              <div className="markdown">
                <ReactMarkdown>{selectedMarkdown}</ReactMarkdown>
              </div>
            </>
          )}
        </div>
      </section>
    </div>
  );
}

