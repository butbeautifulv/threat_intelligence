'use client';

import dynamic from 'next/dynamic';
import { useEffect, useMemo, useRef, useState } from 'react';
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
  const [freezeLayout, setFreezeLayout] = useState(true);
  const [collapsedKinds, setCollapsedKinds] = useState<Record<string, boolean>>({});

  // resizable panes
  const wrapRef = useRef<HTMLDivElement | null>(null);
  const [leftW, setLeftW] = useState(340);
  const [rightW, setRightW] = useState(440);
  const dragRef = useRef<null | { which: 'left' | 'right'; startX: number; startLeft: number; startRight: number }>(
    null
  );

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

  // Initialize kind folders as collapsed by default.
  useEffect(() => {
    const kinds = Array.from(new Set(graph.nodes.map((n) => n.kind || n.labels?.[0] || 'Node'))).sort();
    setCollapsedKinds((prev) => {
      const next: Record<string, boolean> = { ...prev };
      for (const k of kinds) {
        if (next[k] === undefined) next[k] = true;
      }
      return next;
    });
  }, [graph.nodes]);

  // Apply persisted positions (simple localStorage cache).
  useEffect(() => {
    try {
      const raw = localStorage.getItem('ti:graph:pos:v1');
      if (!raw) return;
      const parsed = JSON.parse(raw) as Record<string, { x: number; y: number }>;
      setGraph((g) => ({
        ...g,
        nodes: g.nodes.map((n: any) => {
          const p = parsed[n.id];
          if (!p) return n;
          return { ...n, x: p.x, y: p.y, fx: freezeLayout ? p.x : undefined, fy: freezeLayout ? p.y : undefined };
        }),
      }));
    } catch {
      // ignore
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    function onMove(e: MouseEvent) {
      const d = dragRef.current;
      if (!d) return;
      const el = wrapRef.current;
      if (!el) return;
      const total = el.getBoundingClientRect().width;
      const dx = e.clientX - d.startX;
      const minLeft = 240;
      const minRight = 320;
      if (d.which === 'left') {
        const next = Math.max(minLeft, Math.min(total - minRight - 200, d.startLeft + dx));
        setLeftW(next);
      } else {
        const next = Math.max(minRight, Math.min(total - minLeft - 200, d.startRight - dx));
        setRightW(next);
      }
    }
    function onUp() {
      dragRef.current = null;
    }
    window.addEventListener('mousemove', onMove);
    window.addEventListener('mouseup', onUp);
    return () => {
      window.removeEventListener('mousemove', onMove);
      window.removeEventListener('mouseup', onUp);
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
    const base = graph.nodes.filter((n) => {
      if (!s) return true;
      return `${n.title} ${n.id} ${n.labels.join(' ')} ${n.kind || ''}`.toLowerCase().includes(s);
    });
    return base.slice(0, 400);
  }, [graph.nodes, q]);

  const kindFolders = useMemo(() => {
    const groups = new Map<string, GraphNode[]>();
    for (const n of filteredNodes) {
      const key = n.kind || n.labels?.[0] || 'Node';
      const arr = groups.get(key) || [];
      arr.push(n);
      groups.set(key, arr);
    }
    const sorted = Array.from(groups.entries()).sort((a, b) => b[1].length - a[1].length);
    return sorted.map(([k, nodes]) => [k, nodes.slice(0, 200)] as const);
  }, [filteredNodes]);

  const graphForViz = useMemo(() => {
    const nodeSet = new Set(filteredNodes.map((n) => n.id));
    const edges = graph.edges.filter((e) => nodeSet.has(e.source) && nodeSet.has(e.target));
    return { nodes: filteredNodes, links: edges.map((e) => ({ ...e, source: e.source, target: e.target })) };
  }, [filteredNodes, graph.edges]);

  return (
    <div
      className="app"
      ref={wrapRef}
      style={{
        gridTemplateColumns: `${leftW}px 10px 1fr 10px ${rightW}px`,
      }}
    >
      <section className="card">
        <div className="cardHeader">
          <div className="title">Nodes</div>
          <div className="toolbar">
            <span className="pill">{loading ? 'loading…' : `${graph.nodes.length} total`}</span>
            <button
              className="btn"
              onClick={() => {
                setCollapsedKinds((prev) => {
                  const next: Record<string, boolean> = { ...prev };
                  for (const [k] of kindFolders) next[k] = false;
                  return next;
                });
              }}
            >
              Expand all
            </button>
            <button
              className="btn"
              onClick={() => {
                setCollapsedKinds((prev) => {
                  const next: Record<string, boolean> = { ...prev };
                  for (const [k] of kindFolders) next[k] = true;
                  return next;
                });
              }}
            >
              Collapse all
            </button>
          </div>
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
            {kindFolders.map(([kind, nodes]) => {
              const collapsed = collapsedKinds[kind] !== false; // default collapsed
              return (
                <div key={kind} style={{ marginBottom: 10 }}>
                  <div
                    className="folderHeader"
                    onClick={() => setCollapsedKinds((prev) => ({ ...prev, [kind]: !collapsed }))}
                    role="button"
                    tabIndex={0}
                  >
                    <div className="folderLeft">
                      <span className="caret">{collapsed ? '▶' : '▼'}</span>
                      <div className="folderName">{kind}</div>
                      <span className="pill">{nodes.length}</span>
                    </div>
                  </div>
                  {!collapsed && (
                    <div className="folderBody" style={{ marginTop: 8 }}>
                      <div className="list">
                        {nodes.map((n) => (
                          <div
                            key={n.id}
                            className={`listItem ${selectedId === n.id ? 'listItemActive' : ''}`}
                            onClick={() => setSelectedId(n.id)}
                            role="button"
                            tabIndex={0}
                          >
                            <div className="name">{nodeTitle(n)}</div>
                            <div className="meta">
                              <span className="pill">{n.kind || n.labels[0] || 'Node'}</span>
                              <span className="pill">{n.id}</span>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
          {!filteredNodes.length && <div className="hint">No matches.</div>}
        </div>
      </section>

      <div
        className="resizer"
        onMouseDown={(e) => {
          dragRef.current = { which: 'left', startX: e.clientX, startLeft: leftW, startRight: rightW };
        }}
        role="separator"
        aria-label="Resize left pane"
      />

      <section className="card">
        <div className="cardHeader">
          <div className="title">Graph</div>
          <div className="toolbar">
            <button className={`btn ${freezeLayout ? 'btnActive' : ''}`} onClick={() => setFreezeLayout((v) => !v)}>
              {freezeLayout ? 'Frozen' : 'Free'}
            </button>
            <span className="pill">{filteredNodes.length} shown</span>
          </div>
        </div>
        <div style={{ height: 'calc(100% - 49px)' }}>
          <ForceGraph2D
            graphData={graphForViz as any}
            backgroundColor="rgba(0,0,0,0)"
            nodeLabel={(n: any) => nodeTitle(n as GraphNode)}
            nodeAutoColorBy={(n: any) => (n.kind ?? n.labels?.[0] ?? 'Node')}
            linkColor={() => 'rgba(255,255,255,.20)'}
            linkWidth={1}
            onNodeClick={(n: any) => setSelectedId((n as GraphNode).id)}
            cooldownTicks={freezeLayout ? 120 : 0}
            d3AlphaDecay={freezeLayout ? 0.03 : 0.02}
            enableNodeDrag
            onNodeDragEnd={(n: any) => {
              if (!freezeLayout) return;
              n.fx = n.x;
              n.fy = n.y;
              try {
                const raw = localStorage.getItem('ti:graph:pos:v1') || '{}';
                const parsed = JSON.parse(raw) as Record<string, { x: number; y: number }>;
                parsed[n.id] = { x: n.x, y: n.y };
                localStorage.setItem('ti:graph:pos:v1', JSON.stringify(parsed));
              } catch {
                // ignore
              }
            }}
          />
        </div>
      </section>

      <div
        className="resizer"
        onMouseDown={(e) => {
          dragRef.current = { which: 'right', startX: e.clientX, startLeft: leftW, startRight: rightW };
        }}
        role="separator"
        aria-label="Resize right pane"
      />

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

