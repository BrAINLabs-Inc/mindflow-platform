import React, { memo } from 'react';
import Svg, { Circle, Defs, Ellipse, G, LinearGradient, Path, Polygon, Rect, Stop } from 'react-native-svg';
import { Colors } from '../constants/colors';

export interface ScenePoint {
    x: number;
    y: number;
}

interface RoadmapSceneProps {
    /** Full drawing width (the map plus the side bleed to the screen edges). */
    width: number;
    height: number;
    /** Centre of each roadmap stop, in scene coordinates, in step order. */
    points: ScenePoint[];
    /** Per-step completion, same order as points. */
    completed: boolean[];
    /** Left edge of the map area within the scene (the side bleed width). */
    offsetX: number;
}

// Flat cartoon palette for the scenery (decor only; UI colours stay in Colors).
const P = {
    grassTop: '#E4F7EC',
    grassBottom: '#CDEFDB',
    grassPatch: '#BFE8CF',
    hillFar: '#D6F1E1',
    roadEdge: '#E7CFA2',
    road: '#F7E9CC',
    roadShadow: '#1E2A3A',
    leafLight: '#7DD8A9',
    leaf: '#3FBF86',
    leafDark: '#2A9E6B',
    pine: '#2E9F6E',
    pineDark: '#1F8259',
    trunk: '#B98559',
    trunkDark: '#9C6B42',
    water: '#9AD3F5',
    waterLight: '#C9E9FB',
    rock: '#C9CED8',
    rockDark: '#AEB5C2',
    cloud: '#FFFFFF',
    bird: '#5A6578',
    flagPole: '#6B7280',
};

const FLOWER_COLORS = ['#FF8A3D', '#F25C54', '#F5B523', '#6C63FF', '#FF9EC4'];

/** Smooth S-curve through the stops (same shape the old dashed path used). */
function roadPath(points: ScenePoint[], width: number, offsetX: number): string[] {
    const segs: string[] = [];
    for (let i = 0; i < points.length - 1; i++) {
        const a = points[i];
        const b = points[i + 1];
        const goingLeft = b.x < a.x;
        const c1x = offsetX + (width - 2 * offsetX) * (goingLeft ? 0.4 : 0.6);
        const c2x = offsetX + (width - 2 * offsetX) * (goingLeft ? 0.6 : 0.4);
        segs.push(`M ${a.x} ${a.y} C ${c1x} ${a.y}, ${c2x} ${b.y}, ${b.x} ${b.y}`);
    }
    return segs;
}

function RoundTree({ x, y, s = 1 }: { x: number; y: number; s?: number }) {
    return (
        <G transform={`translate(${x} ${y}) scale(${s})`}>
            <Ellipse cx={0} cy={2} rx={18} ry={5} fill={P.roadShadow} opacity={0.08} />
            <Rect x={-3.5} y={-22} width={7} height={24} rx={3} fill={P.trunk} />
            <Rect x={0} y={-22} width={3.5} height={24} rx={1.5} fill={P.trunkDark} opacity={0.5} />
            <Circle cx={-9} cy={-30} r={13} fill={P.leaf} />
            <Circle cx={9} cy={-30} r={13} fill={P.leafDark} />
            <Circle cx={0} cy={-42} r={15} fill={P.leaf} />
            <Circle cx={-5} cy={-46} r={6} fill={P.leafLight} opacity={0.9} />
        </G>
    );
}

function PineTree({ x, y, s = 1 }: { x: number; y: number; s?: number }) {
    return (
        <G transform={`translate(${x} ${y}) scale(${s})`}>
            <Ellipse cx={0} cy={2} rx={14} ry={4} fill={P.roadShadow} opacity={0.08} />
            <Rect x={-3} y={-12} width={6} height={14} rx={2} fill={P.trunk} />
            <Polygon points="0,-58 -16,-28 16,-28" fill={P.pine} />
            <Polygon points="0,-46 -19,-12 19,-12" fill={P.pine} />
            <Polygon points="0,-58 0,-28 16,-28" fill={P.pineDark} opacity={0.55} />
            <Polygon points="0,-46 0,-12 19,-12" fill={P.pineDark} opacity={0.55} />
        </G>
    );
}

function Bush({ x, y, s = 1 }: { x: number; y: number; s?: number }) {
    return (
        <G transform={`translate(${x} ${y}) scale(${s})`}>
            <Ellipse cx={0} cy={1} rx={20} ry={4} fill={P.roadShadow} opacity={0.07} />
            <Circle cx={-11} cy={-7} r={9} fill={P.leafDark} />
            <Circle cx={10} cy={-7} r={9} fill={P.leafDark} />
            <Circle cx={0} cy={-12} r={11} fill={P.leaf} />
            <Circle cx={-4} cy={-16} r={4} fill={P.leafLight} />
        </G>
    );
}

function Flower({ x, y, color }: { x: number; y: number; color: string }) {
    return (
        <G transform={`translate(${x} ${y})`}>
            <Path d="M0 0 L0 -8" stroke={P.leafDark} strokeWidth={1.5} strokeLinecap="round" />
            <Circle cx={-2.6} cy={-9} r={2.4} fill={color} />
            <Circle cx={2.6} cy={-9} r={2.4} fill={color} />
            <Circle cx={0} cy={-11.6} r={2.4} fill={color} />
            <Circle cx={0} cy={-6.6} r={2.4} fill={color} />
            <Circle cx={0} cy={-9} r={1.6} fill="#FFF3B0" />
        </G>
    );
}

function FlowerPatch({ x, y, seed = 0 }: { x: number; y: number; seed?: number }) {
    const offs = [[0, 0], [9, 3], [-8, 4], [4, 9]];
    return (
        <G>
            {offs.map(([dx, dy], i) => (
                <Flower key={i} x={x + dx} y={y + dy} color={FLOWER_COLORS[(i + seed) % FLOWER_COLORS.length]} />
            ))}
        </G>
    );
}

function Rocks({ x, y }: { x: number; y: number }) {
    return (
        <G transform={`translate(${x} ${y})`}>
            <Ellipse cx={2} cy={2} rx={20} ry={4} fill={P.roadShadow} opacity={0.07} />
            <Path d="M-16 0 C-16 -10 -6 -14 0 -12 C6 -10 8 -4 8 0 Z" fill={P.rock} />
            <Path d="M-4 -11 C2 -10 8 -4 8 0 L0 0 Z" fill={P.rockDark} opacity={0.6} />
            <Path d="M6 0 C6 -6 12 -8 16 -6 C19 -4 19 -2 19 0 Z" fill={P.rockDark} />
        </G>
    );
}

function Pond({ x, y }: { x: number; y: number }) {
    return (
        <G transform={`translate(${x} ${y})`}>
            <Ellipse cx={0} cy={0} rx={38} ry={15} fill={P.water} />
            <Ellipse cx={-6} cy={-3} rx={24} ry={7} fill={P.waterLight} />
            <Path d="M-14 -2 L-4 -2" stroke="#FFFFFF" strokeWidth={2} strokeLinecap="round" />
            <Path d="M4 4 L14 4" stroke="#FFFFFF" strokeWidth={2} strokeLinecap="round" opacity={0.8} />
            {/* reeds */}
            <Path d="M32 -4 L30 -20 M36 -3 L37 -17 M28 -2 L24 -14" stroke={P.leafDark} strokeWidth={2} strokeLinecap="round" />
        </G>
    );
}

function Cloud({ x, y, s = 1 }: { x: number; y: number; s?: number }) {
    return (
        <G transform={`translate(${x} ${y}) scale(${s})`} opacity={0.95}>
            <Ellipse cx={0} cy={0} rx={16} ry={7} fill={P.cloud} />
            <Circle cx={-6} cy={-4} r={7} fill={P.cloud} />
            <Circle cx={5} cy={-6} r={9} fill={P.cloud} />
        </G>
    );
}

function Birds({ x, y }: { x: number; y: number }) {
    return (
        <G transform={`translate(${x} ${y})`}>
            <Path d="M0 0 Q4 -4 8 0 Q12 -4 16 0" stroke={P.bird} strokeWidth={1.6} fill="none" strokeLinecap="round" />
            <Path d="M20 -8 Q23 -11 26 -8 Q29 -11 32 -8" stroke={P.bird} strokeWidth={1.4} fill="none" strokeLinecap="round" />
        </G>
    );
}

function FinishFlag({ x, y }: { x: number; y: number }) {
    return (
        <G transform={`translate(${x} ${y})`}>
            <Ellipse cx={0} cy={2} rx={14} ry={4} fill={P.roadShadow} opacity={0.1} />
            <Rect x={-1.5} y={-44} width={3} height={46} rx={1.5} fill={P.flagPole} />
            <Path d="M1.5 -44 L28 -37 L1.5 -29 Z" fill={Colors.accent} />
            <Path d="M1.5 -44 L28 -37 L1.5 -37 Z" fill="#FFFFFF" opacity={0.25} />
            <Circle cx={0} cy={-45} r={3} fill={Colors.ochre} />
        </G>
    );
}

// Original illustrated "journey map": a winding road through a cartoon meadow,
// with the five roadmap stops placed on it by the dashboard. Scenery only — the
// interactive stops/labels are rendered on top by DashboardScreen. Deliberately
// neutral (landscape, no meditation imagery), so it suits both study arms.
function RoadmapSceneImpl({ width: W, height: H, points, completed, offsetX }: RoadmapSceneProps) {
    const segments = roadPath(points, W, offsetX);
    const last = points[points.length - 1];
    const finish = { x: W * 0.5, y: H - 46 };
    const tail = last
        ? `M ${last.x} ${last.y} C ${last.x} ${last.y + 50}, ${W * 0.66} ${finish.y}, ${finish.x} ${finish.y}`
        : '';
    const allDone = completed.length > 0 && completed.every(Boolean);
    const p = points;
    const left = offsetX;
    const right = W - offsetX;

    const road = (d: string, key: string, doneTrail: boolean) => (
        <G key={key}>
            <Path d={d} stroke={P.roadShadow} strokeOpacity={0.07} strokeWidth={30} fill="none" strokeLinecap="round" transform="translate(0 3)" />
            <Path d={d} stroke={P.roadEdge} strokeWidth={28} fill="none" strokeLinecap="round" />
            <Path d={d} stroke={P.road} strokeWidth={21} fill="none" strokeLinecap="round" />
            <Path
                d={d}
                stroke={doneTrail ? Colors.primary : '#FFFFFF'}
                strokeWidth={doneTrail ? 3.5 : 3}
                strokeDasharray="9 9"
                fill="none"
                strokeLinecap="round"
            />
        </G>
    );

    return (
        <Svg width={W} height={H} viewBox={`0 0 ${W} ${H}`}>
            <Defs>
                <LinearGradient id="meadow" x1="0" y1="0" x2="0" y2="1">
                    <Stop offset="0" stopColor={P.grassTop} />
                    <Stop offset="1" stopColor={P.grassBottom} />
                </LinearGradient>
            </Defs>

            {/* Rolling meadow with a soft far hill along the top edge. */}
            <Path d={`M0 26 C ${W * 0.2} 6, ${W * 0.42} 30, ${W * 0.62} 16 S ${W * 0.9} 10, ${W} 22 L ${W} ${H} L 0 ${H} Z`} fill={P.hillFar} />
            {/* Meadow with a curved shoreline at the bottom (the panel's wave below reads as water). */}
            <Path
                d={`M0 40 C ${W * 0.25} 22, ${W * 0.5} 48, ${W * 0.75} 30 S ${W} 36, ${W} 36 L ${W} ${H - 26} C ${W * 0.78} ${H - 2}, ${W * 0.55} ${H - 30}, ${W * 0.3} ${H - 12} S ${W * 0.06} ${H - 18}, 0 ${H - 10} Z`}
                fill="url(#meadow)"
            />
            <Path
                d={`M 0 ${H - 10} C ${W * 0.06} ${H - 18}, ${W * 0.2} ${H - 18}, ${W * 0.3} ${H - 12} S ${W * 0.78} ${H - 2}, ${W} ${H - 26}`}
                stroke={P.roadEdge}
                strokeWidth={5}
                fill="none"
                strokeLinecap="round"
                opacity={0.7}
            />
            <Ellipse cx={W * 0.25} cy={H * 0.42} rx={W * 0.32} ry={34} fill={P.grassPatch} opacity={0.45} />
            <Ellipse cx={W * 0.8} cy={H * 0.7} rx={W * 0.3} ry={30} fill={P.grassPatch} opacity={0.45} />

            {/* Sky details above the hills. */}
            <Cloud x={left + 24} y={14} s={0.9} />
            <Cloud x={W * 0.58} y={10} s={0.7} />
            <Birds x={W * 0.3} y={16} />

            {/* The road, segment by segment: the centre line turns green once the
                step it leaves from is completed, like a trail behind you. */}
            {segments.map((d, i) => road(d, `seg${i}`, !!completed[i]))}
            {tail ? road(tail, 'tail', allDone) : null}

            {/* Ground shadows under each stop. */}
            {p.map((pt, i) => (
                <Ellipse key={`sh${i}`} cx={pt.x} cy={pt.y + 30} rx={24} ry={6} fill={P.roadShadow} opacity={0.1} />
            ))}

            {/* Scenery in the free space beside each stop (stops alternate right/left). */}
            {p[0] && <RoundTree x={left + 26} y={p[0].y + 34} s={0.95} />}
            {p[0] && <Bush x={left + 64} y={p[0].y + 36} s={0.75} />}
            {p[1] && <Pond x={right - 38} y={p[1].y + 18} />}
            {p[1] && <FlowerPatch x={right - 90} y={p[1].y + 44} seed={1} />}
            {p[2] && <PineTree x={left + 20} y={p[2].y + 32} s={0.95} />}
            {p[2] && <PineTree x={left + 52} y={p[2].y + 38} s={0.8} />}
            {p[2] && <FlowerPatch x={left + 30} y={p[2].y - 40} seed={2} />}
            {p[3] && <Rocks x={right - 44} y={p[3].y + 30} />}
            {p[3] && <Bush x={right - 22} y={p[3].y - 8} s={0.85} />}
            {p[4] && <RoundTree x={left + 30} y={p[4].y + 30} s={0.85} />}
            {p[4] && <FlowerPatch x={left + 68} y={p[4].y + 34} seed={3} />}

            <FinishFlag x={finish.x} y={finish.y} />
        </Svg>
    );
}

export const RoadmapScene = memo(RoadmapSceneImpl);
export default RoadmapScene;
