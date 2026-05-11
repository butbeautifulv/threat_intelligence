'use client';

import dynamic from 'next/dynamic';
import { useEffect, useMemo, useRef, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import type { GraphData, GraphNode } from './types';
import { forceCollide } from 'd3-force';

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
  const graphWrapRef = useRef<HTMLDivElement | null>(null);
  const nodeDragRef = useRef<{ active: boolean; id: string | number | null }>({ active: false, id: null });
  const [graph, setGraph] = useState<GraphData>({ nodes: [], edges: [] });
  const [loading, setLoading] = useState(true);
  const [q, setQ] = useState('');
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [hoveredId, setHoveredId] = useState<string | null>(null);
  const [selected, setSelected] = useState<GraphNode | null>(null);
  const [selectedMarkdown, setSelectedMarkdown] = useState<string>('');
  const [collapsedKinds, setCollapsedKinds] = useState<Record<string, boolean>>({});

  // resizable panes
  const wrapRef = useRef<HTMLDivElement | null>(null);
  const [leftW, setLeftW] = useState(340);
  const [rightW, setRightW] = useState(440);
  const dragRef = useRef<null | { which: 'left' | 'right'; startX: number; startLeft: number; startRight: number }>(
    null
  );

  const [graphSize, setGraphSize] = useState({ w: 0, h: 0 });

  // canvas “Obsidian-like” knobs
  const [nodeSize, setNodeSize] = useState(5);
  const [chargeStrength, setChargeStrength] = useState(-55);
  const [linkDistance, setLinkDistance] = useState(32);
  const [clampRadius, setClampRadius] = useState(650);
  const [dragRadius, setDragRadius] = useState(220);

  const focusNode = (id: string) => {
    const fg = fgRef.current;
    if (!fg) return;
    try {
      const data = fg.graphData?.();
      const nodes: any[] = data?.nodes || [];
      const n = nodes.find((x) => x?.id === id);
      if (!n || typeof n.x !== 'number' || typeof n.y !== 'number') return;
      // Run after layout/canvas size settles (esp. after pane resize).
      requestAnimationFrame(() => {
        try {
          fg.centerAt(n.x, n.y, 300);
          const z = typeof fg.zoom === 'function' ? fg.zoom() : 1;
          if (typeof z === 'number' && z < 2.2) fg.zoom(2.2, 300);
        } catch {
          // ignore
        }
      });
    } catch {
      // ignore
    }
  };

  // Keep canvas stable when panes are hidden/resized.
  useEffect(() => {
    const el = graphWrapRef.current;
    if (!el) return;
    const ro = new ResizeObserver(() => {
      const r = el.getBoundingClientRect();
      const w = Math.max(0, Math.floor(r.width));
      const h = Math.max(0, Math.floor(r.height));
      setGraphSize((prev) => (prev.w === w && prev.h === h ? prev : { w, h }));
    });
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

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

  // (No persisted/frozen layout controls — keep interaction simple like Obsidian.)

  useEffect(() => {
    function onMove(e: MouseEvent) {
      const d = dragRef.current;
      if (!d) return;
      const dx = e.clientX - d.startX;
      const minLeft = 260;
      const minRight = 320;
      if (d.which === 'left') {
        const next = Math.max(minLeft, d.startLeft + dx);
        setLeftW(next);
      } else {
        const next = Math.max(minRight, d.startRight - dx);
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
    return base;
  }, [graph.nodes, q]);

  const browserNodes = useMemo(() => {
    // Don't show synthetic category hubs as a “kind” folder in the browser.
    return filteredNodes.filter((n) => !n.labels?.includes('Category'));
  }, [filteredNodes]);

  const kindFolders = useMemo(() => {
    const groups = new Map<string, GraphNode[]>();
    for (const n of browserNodes) {
      const key = n.kind || n.labels?.[0] || 'Node';
      const arr = groups.get(key) || [];
      arr.push(n);
      groups.set(key, arr);
    }
    const sorted = Array.from(groups.entries()).sort((a, b) => b[1].length - a[1].length);
    return sorted.map(([k, nodes]) => [k, nodes.slice(0, 200)] as const);
  }, [browserNodes]);

  const allKinds = useMemo(() => {
    return Array.from(new Set(browserNodes.map((n) => n.kind || n.labels?.[0] || 'Node'))).sort();
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

  const activeId = hoveredId || selectedId;
  const activeNeighborhood = useMemo(() => {
    if (!activeId) return null;
    const neigh = adjacency.get(activeId) || new Set<string>();
    return { id: activeId, neigh, isHover: Boolean(hoveredId) };
  }, [adjacency, activeId, hoveredId]);

  const clearActive = () => {
    setSelectedId(null);
    setHoveredId(null);
    setSelected(null);
    setSelectedMarkdown('');
  };

  // Tune forces (safe to update; does not recreate graph).
  useEffect(() => {
    const t = window.setTimeout(() => {
      const fg = fgRef.current;
      if (!fg?.d3Force) return;
      try {
        fg.d3Force('charge')?.strength(chargeStrength);
        fg.d3Force('link')?.distance(linkDistance);
        fg.d3Force(
          'collide',
          forceCollide((n: any) => {
            // Category hubs are larger; nodes repel within their radius (Obsidian-like).
            const base = n?.labels?.includes('Category') ? 30 : 18;
            return base + nodeSize * 1.2;
          }).strength(0.8)
        );
      } catch {
        // ignore
      }
    }, 0);
    return () => window.clearTimeout(t);
  }, [chargeStrength, linkDistance, nodeSize]);

  // When selecting from browser, center/zoom to node position.
  useEffect(() => {
    if (!selectedId) return;
    focusNode(selectedId);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedId]);

  // Re-focus after pane resize (canvas size changes can shift perceived center).
  useEffect(() => {
    if (!selectedId) return;
    const t = window.setTimeout(() => focusNode(selectedId), 50);
    return () => window.clearTimeout(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [leftW, rightW, selectedId]);

  return (
    <div
      className="app"
      ref={wrapRef}
      style={{
        gridTemplateColumns: `${leftW}px 10px 1fr 10px ${rightW}px`,
      }}
    >
      <section className="card" style={{ minHeight: 0 }}>
        <div className="cardHeader">
          <div className="title">Nodes</div>
          <div className="toolbar toolbarWrap">
            <span className="pill">{loading ? 'loading…' : `${graph.nodes.length}`}</span>
            <button
              className="btn iconBtn"
              onClick={() => {
                setCollapsedKinds((prev) => {
                  const next: Record<string, boolean> = { ...prev };
                  for (const k of allKinds) next[k] = false;
                  return next;
                });
              }}
              title="Expand all kinds"
            >
              +
            </button>
            <button
              className="btn iconBtn"
              onClick={() => {
                setCollapsedKinds((prev) => {
                  const next: Record<string, boolean> = { ...prev };
                  for (const k of allKinds) next[k] = true;
                  return next;
                });
              }}
              title="Collapse all kinds"
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
                    onClick={() => {
                      const nextCollapsed = !collapsed;
                      setCollapsedKinds((prev) => ({ ...prev, [kind]: nextCollapsed }));
                      // When expanding a folder, highlight the central hub node for that kind.
                      if (!nextCollapsed) {
                        setSelectedId(`category:${kind}`);
                      }
                    }}
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

      <section className="card" style={{ minHeight: 0 }}>
        <div className="cardHeader">
          <div className="title">Graph</div>
          <div className="toolbar toolbarWrap">
            <button className="btn iconBtn" onClick={clearActive} title="Reset">
              <svg className="icon" viewBox="0 0 24 24" aria-hidden="true">
                <path d="M12 6v12M6 12h12" />
              </svg>
            </button>
            <div className="sliderRow">
              <span>node</span>
              <input
                className="slider"
                type="range"
                min={3}
                max={10}
                value={nodeSize}
                onChange={(e) => setNodeSize(Number(e.target.value))}
              />
              <span>{nodeSize}</span>
            </div>
          </div>
        </div>
        <div ref={graphWrapRef} style={{ height: 'calc(100% - 49px)' }}>
          <ForceGraph2D
            ref={fgRef}
            graphData={graphForViz as any}
            width={graphSize.w || undefined}
            height={graphSize.h || undefined}
            backgroundColor="rgba(0,0,0,0)"
            nodeLabel={(n: any) => nodeTitle(n as GraphNode)}
            nodeColor={(n: any) => {
              const labels: string[] = n.labels || [];
              if (labels.includes('Category')) {
                if (!activeNeighborhood) return COLOR_NODE;
                if (n.id === activeNeighborhood.id) return COLOR_ACCENT;
                if (activeNeighborhood.neigh.has(n.id)) return COLOR_NODE; // keep neighbor standard
                return COLOR_NODE_DIM;
              }
              if (!activeNeighborhood) return COLOR_NODE;
              if (n.id === activeNeighborhood.id) return COLOR_ACCENT;
              if (activeNeighborhood.neigh.has(n.id)) return COLOR_NODE; // keep neighbor standard
              return COLOR_NODE_DIM;
            }}
            linkColor={(l: any) => {
              if (!activeNeighborhood) return COLOR_EDGE;
              const s = typeof l.source === 'string' ? l.source : l.source?.id;
              const t = typeof l.target === 'string' ? l.target : l.target?.id;
              const sel = activeNeighborhood.id;
              if (s === sel || t === sel) return COLOR_ACCENT; // active edges purple
              if (activeNeighborhood.neigh.has(String(s)) && activeNeighborhood.neigh.has(String(t))) return 'rgba(255,255,255,.09)';
              return COLOR_EDGE_DIM;
            }}
            linkWidth={() => 0.6}
            onNodeClick={(n: any) => setSelectedId((n as GraphNode).id)}
            onNodeHover={(n: any) => setHoveredId(n?.id ?? null)}
            onBackgroundClick={() => {
              clearActive();
            }}
            d3AlphaDecay={0.02}
            d3VelocityDecay={0.35}
            nodeRelSize={nodeSize}
            nodeVal={(n: any) => (n.labels?.includes('Category') ? 10 : 1)}
            nodeCanvasObjectMode={() => 'after'}
            nodeCanvasObject={(n: any, ctx: CanvasRenderingContext2D, globalScale: number) => {
              const x = n.x as number;
              const y = n.y as number;
              const label = String(n.title || n.id);

              // Category hub tiny centered label
              if (n.labels?.includes('Category')) {
                const fontSize = Math.max(6.5, Math.min(9, 9 / globalScale));
                ctx.font = `800 ${fontSize}px ui-sans-serif, system-ui`;
                const isActive = activeNeighborhood && n.id === activeNeighborhood.id;
                ctx.fillStyle = isActive ? 'rgba(255,255,255,.95)' : 'rgba(0,0,0,.85)';
                ctx.textAlign = 'center';
                ctx.textBaseline = 'middle';
                const maxChars = 10;
                const text = label.length > maxChars ? `${label.slice(0, maxChars - 1)}…` : label;
                ctx.fillText(text, x, y);
                return;
              }

              // Node labels when zoomed in
              // Avoid messy overlaps: show labels only for active neighborhood,
              // unless we're extremely zoomed in.
              const showAll = globalScale >= 5;
              if (!showAll) {
                if (!activeNeighborhood) return;
                if (!(n.id === activeNeighborhood.id || activeNeighborhood.neigh.has(n.id))) return;
              }

              if (globalScale < 2.8) return;
              const fontSize = Math.max(7, Math.min(11, 11 / globalScale));
              ctx.font = `600 ${fontSize}px ui-sans-serif, system-ui`;
              ctx.fillStyle = 'rgba(230,237,247,1)';
              ctx.textAlign = 'center';
              ctx.textBaseline = 'top';
              const maxChars = 22;
              const text = label.length > maxChars ? `${label.slice(0, maxChars - 1)}…` : label;
              const ty = y + nodeSize * 2.6;
              const tw = ctx.measureText(text).width;
              const padX = 6;
              const padY = 3;
              ctx.fillStyle = 'rgba(11,15,23,.88)';
              ctx.fillRect(x - tw / 2 - padX, ty - 1, tw + padX * 2, fontSize + padY * 2);
              ctx.fillStyle = 'rgba(230,237,247,1)';
              ctx.fillText(text, x, ty + padY);
            }}
            onEngineTick={() => {
              // Soft clamp to keep nodes near center (prevents “flyaway” beyond viewport).
              const nodes = graphForViz.nodes as any[];
              const R = clampRadius;
              for (const n of nodes) {
                if (!n || n.labels?.includes('Category')) continue;
                if (typeof n.x !== 'number' || typeof n.y !== 'number') continue;
                const d = Math.hypot(n.x, n.y);
                if (d > R) {
                  const s = R / d;
                  n.x *= s;
                  n.y *= s;
                }
              }
            }}
            enableNodeDrag
            onNodeDrag={(n: any) => {
              // react-force-graph doesn't expose onNodeDragStart in its TS types.
              // Emulate "start" on the first onNodeDrag callback of a drag session.
              const nid = (n?.id ?? null) as string | number | null;
              if (!nodeDragRef.current.active || nodeDragRef.current.id !== nid) {
                nodeDragRef.current = { active: true, id: nid };
                const fg = fgRef.current;
                try {
                  fg?.d3AlphaTarget?.(0.2);
                  fg?.d3ReheatSimulation?.();
                } catch {
                  // ignore
                }
              }
              // Keep the dragged node pinned to cursor. No “neighbor gravity” here.
              n.fx = n.x;
              n.fy = n.y;
            }}
            onNodeDragEnd={(n: any) => {
              nodeDragRef.current = { active: false, id: null };
              const fg = fgRef.current;
              try {
                fg?.d3AlphaTarget?.(0);
              } catch {
                // ignore
              }
              n.fx = null;
              n.fy = null;
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

      <section className="card right" style={{ minHeight: 0 }}>
        <div className="cardHeader">
          <div className="title">Markdown</div>
          <div className="toolbar">
            {selected ? <span className="pill">{selected.labels[0] ?? 'Node'}</span> : <span className="pill">none</span>}
          </div>
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

