'use client';

import dynamic from 'next/dynamic';
import { useEffect, useMemo, useRef, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import type { GraphData, GraphNode } from './types';

const ForceGraph2D = dynamic(() => import('react-force-graph-2d'), { ssr: false });

function nodeTitle(n: GraphNode) {
  return n.title || n.id;
}

const COLOR_NODE = 'rgba(230,237,247,.85)'; // default white-ish
const COLOR_NODE_DIM = 'rgba(230,237,247,.14)';
const COLOR_NODE_NEIGH = 'rgba(230,237,247,.55)';
const COLOR_EDGE = 'rgba(255,255,255,.12)';
const COLOR_EDGE_DIM = 'rgba(255,255,255,.05)';
const COLOR_ACCENT = 'rgba(122,162,247,.95)';
const COLOR_ACCENT_DIM = 'rgba(122,162,247,.30)';

export default function GraphExplorer() {
  const fgRef = useRef<any>(null);
  const [graph, setGraph] = useState<GraphData>({ nodes: [], edges: [] });
  const [loading, setLoading] = useState(true);
  const [q, setQ] = useState('');
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [selected, setSelected] = useState<GraphNode | null>(null);
  const [selectedMarkdown, setSelectedMarkdown] = useState<string>('');
  const [freezeLayout, setFreezeLayout] = useState(true);
  const [collapsedKinds, setCollapsedKinds] = useState<Record<string, boolean>>({});
  const dragMoveRef = useRef<{ id: string; lastX: number; lastY: number } | null>(null);

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
        if (!cancelled) {
          // Pin category hubs near the center (Obsidian-like “clusters”).
          const cats = data.nodes.filter((n) => n.labels?.includes('Category'));
          const angleStep = (Math.PI * 2) / Math.max(1, cats.length);
          const catPos = new Map<string, { x: number; y: number }>();
          cats.forEach((n, idx) => {
            const a = idx * angleStep;
            const r = 70;
            catPos.set(n.id, { x: Math.cos(a) * r, y: Math.sin(a) * r });
          });
          setGraph({
            ...data,
            nodes: data.nodes.map((n: any) => {
              if (n.labels?.includes('Category')) {
                const p = catPos.get(n.id) || { x: 0, y: 0 };
                return { ...n, x: p.x, y: p.y, fx: p.x, fy: p.y };
              }
              return n;
            }),
          });
        }
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

  const allKinds = useMemo(() => {
    return Array.from(new Set(graph.nodes.map((n) => n.kind || n.labels?.[0] || 'Node'))).sort();
  }, [graph.nodes]);

  const graphForViz = useMemo(() => {
    const nodeSet = new Set(filteredNodes.map((n) => n.id));
    const edges = graph.edges.filter((e) => nodeSet.has(e.source) && nodeSet.has(e.target));
    return { nodes: filteredNodes, links: edges.map((e) => ({ ...e, source: e.source, target: e.target })) };
  }, [filteredNodes, graph.edges]);

  const adjacency = useMemo(() => {
    const map = new Map<string, Set<string>>();
    const add = (a: string, b: string) => {
      const s = map.get(a) || new Set<string>();
      s.add(b);
      map.set(a, s);
    };
    for (const e of graphForViz.links as any[]) {
      const s = typeof e.source === 'string' ? e.source : e.source?.id;
      const t = typeof e.target === 'string' ? e.target : e.target?.id;
      if (!s || !t) continue;
      add(s, t);
      add(t, s);
    }
    return map;
  }, [graphForViz.links]);

  const selectedNeighborhood = useMemo(() => {
    if (!selectedId) return null;
    const neigh = adjacency.get(selectedId) || new Set<string>();
    return { id: selectedId, neigh };
  }, [adjacency, selectedId]);

  // Tune forces once (do NOT reapply on every render/click).
  useEffect(() => {
    const t = window.setTimeout(() => {
      const fg = fgRef.current;
      if (!fg?.d3Force) return;
      try {
        fg.d3Force('charge')?.strength(-55);
        fg.d3Force('link')?.distance(32);
      } catch {
        // ignore
      }
    }, 0);
    return () => window.clearTimeout(t);
  }, []);

  // When selecting from browser, center/zoom to node position.
  useEffect(() => {
    if (!selectedId) return;
    const fg = fgRef.current;
    if (!fg) return;
    const n = (graphForViz.nodes as any[]).find((x) => x.id === selectedId);
    if (!n || typeof n.x !== 'number' || typeof n.y !== 'number') return;
    try {
      // Keep this subtle to avoid “jerk”: center always; zoom only if far out.
      fg.centerAt(n.x, n.y, 350);
      const z = typeof fg.zoom === 'function' ? fg.zoom() : 1;
      if (typeof z === 'number' && z < 2.2) fg.zoom(2.2, 350);
    } catch {
      // ignore
    }
  }, [selectedId, graphForViz.nodes]);

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
            <span className="pill">{loading ? 'loading…' : `${graph.nodes.length}`}</span>
            <button
              className="btn"
              onClick={() => {
                setCollapsedKinds((prev) => {
                  const next: Record<string, boolean> = { ...prev };
                  for (const k of allKinds) next[k] = false;
                  return next;
                });
              }}
            >
              +
            </button>
            <button
              className="btn"
              onClick={() => {
                setCollapsedKinds((prev) => {
                  const next: Record<string, boolean> = { ...prev };
                  for (const k of allKinds) next[k] = true;
                  return next;
                });
              }}
            >
              −
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
                      <span className="caret">{collapsed ? '+' : '−'}</span>
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
            ref={fgRef}
            graphData={graphForViz as any}
            backgroundColor="rgba(0,0,0,0)"
            nodeLabel={(n: any) => nodeTitle(n as GraphNode)}
            nodeColor={(n: any) => {
              const labels: string[] = n.labels || [];
              if (labels.includes('Category')) return 'rgba(230,237,247,.55)';
              if (!selectedNeighborhood) return COLOR_NODE;
              if (n.id === selectedNeighborhood.id) return COLOR_ACCENT;
              if (selectedNeighborhood.neigh.has(n.id)) return COLOR_NODE_NEIGH;
              return COLOR_NODE_DIM;
            }}
            linkColor={(l: any) => {
              if (!selectedNeighborhood) return COLOR_EDGE;
              const s = typeof l.source === 'string' ? l.source : l.source?.id;
              const t = typeof l.target === 'string' ? l.target : l.target?.id;
              const sel = selectedNeighborhood.id;
              if (s === sel || t === sel) return COLOR_ACCENT_DIM;
              if (selectedNeighborhood.neigh.has(String(s)) && selectedNeighborhood.neigh.has(String(t))) return 'rgba(255,255,255,.09)';
              return COLOR_EDGE_DIM;
            }}
            linkWidth={(l: any) => {
              if (!selectedNeighborhood) return 1;
              const s = typeof l.source === 'string' ? l.source : l.source?.id;
              const t = typeof l.target === 'string' ? l.target : l.target?.id;
              const sel = selectedNeighborhood.id;
              return s === sel || t === sel ? 2 : 1;
            }}
            onNodeClick={(n: any) => setSelectedId((n as GraphNode).id)}
            cooldownTicks={freezeLayout ? 120 : 0}
            d3AlphaDecay={freezeLayout ? 0.03 : 0.02}
            d3VelocityDecay={0.35}
            nodeRelSize={5}
            nodeVal={(n: any) => (n.labels?.includes('Category') ? 4 : 1)}
            nodeCanvasObjectMode={() => 'after'}
            nodeCanvasObject={(n: any, ctx: CanvasRenderingContext2D, globalScale: number) => {
              if (!n.labels?.includes('Category')) return;
              const label = String(n.title || n.id);
              const fontSize = Math.max(10, 16 / globalScale);
              ctx.font = `600 ${fontSize}px ui-sans-serif, system-ui`;
              ctx.fillStyle = 'rgba(230,237,247,.75)';
              ctx.textAlign = 'center';
              ctx.textBaseline = 'middle';
              ctx.fillText(label, n.x, n.y);
            }}
            onEngineTick={() => {
              // Soft clamp to keep nodes near center (prevents “flyaway” beyond viewport).
              const nodes = graphForViz.nodes as any[];
              const R = 650;
              for (const n of nodes) {
                if (!n || n.labels?.includes('Category')) continue;
                if (typeof n.x !== 'number' || typeof n.y !== 'number') continue;
                const d = Math.hypot(n.x, n.y);
                if (d > R) {
                  const s = R / d;
                  n.x *= s;
                  n.y *= s;
                  if (freezeLayout) {
                    n.fx = n.x;
                    n.fy = n.y;
                  }
                }
              }
            }}
            enableNodeDrag
            onNodeDrag={(n: any) => {
              // Obsidian-ish behavior: move nodes by radius with falloff (gravity-like).
              const last = dragMoveRef.current;
              if (!last || last.id !== n.id) {
                dragMoveRef.current = { id: n.id, lastX: n.x, lastY: n.y };
                return;
              }
              const dx = n.x - last.lastX;
              const dy = n.y - last.lastY;
              dragMoveRef.current = { id: n.id, lastX: n.x, lastY: n.y };
              if (!dx && !dy) return;
              const nodes = graphForViz.nodes as any[];
              const R = 220; // influence radius
              for (const nn of nodes) {
                if (!nn || nn.id === n.id || nn.labels?.includes('Category')) continue;
                if (typeof nn.x !== 'number' || typeof nn.y !== 'number') continue;
                const dist = Math.hypot(nn.x - n.x, nn.y - n.y);
                if (dist > R) continue;
                const w = Math.max(0, 1 - dist / R);
                const k = w * w * 0.75; // quadratic falloff
                nn.x += dx * k;
                nn.y += dy * k;
                if (freezeLayout) {
                  nn.fx = nn.x;
                  nn.fy = nn.y;
                }
              }
            }}
            onNodeDragEnd={(n: any) => {
              dragMoveRef.current = null;
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

