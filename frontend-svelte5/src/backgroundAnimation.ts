interface Position {
    x: number;
    y: number;
}

interface ColorPair {
    light: string;
    dark: string;
}

interface RowState {
    type: string;
    color: ColorPair;
    branch: number;
    element: SVGElement;
    index: number;
    nr: number;
}

interface TileColors {
    top: string;
    front: string;
    right: string;
}

interface FloorConfig {
    viewOffsetX: number;
    viewOffsetY: number;
}

interface TileMovement {
    totalDistance: number;
    currentTileProgress: number;
    tileHeight: number;
}

const movement: TileMovement = {
    totalDistance: 0,
    currentTileProgress: 0,
    tileHeight: 12
};

const MOVEMENT_SPEED = 3;

export const COLORS = {
    YELLOW: {
        light: "#ffd54f",
        dark: "#ffb300"
    },
    RED: {
        light: "#e63835",
        dark: "#ca1c19"
    },
    GREEN: {
        light: "#bbda9b",
        dark: "#7cb342"
    },
    BLUE: {
        light: "#4a90e2",
        dark: "#2171d0"
    },
    PURPLE: {
        light: "#9575cd",
        dark: "#5e35b1"
    },
    ORANGE: {
        light: "#ff9800",
        dark: "#f57c00"
    },
    PINK: {
        light: "#ec407a",
        dark: "#d81b60"
    },
    TEAL: {
        light: "#26a69a",
        dark: "#00897b"
    },
    CYAN: {
        light: "#4dd0e1",
        dark: "#00acc1"
    },
    BROWN: {
        light: "#a1887f",
        dark: "#795548"
    }
};

const floorConfig: FloorConfig = {
    viewOffsetX: 468,
    viewOffsetY: 55
};

function createDot(id: number, position: Position, colors: ColorPair): SVGElement {
    const { x, y } = position;
    const { light, dark } = colors;
    const group = document.createElementNS("http://www.w3.org/2000/svg", "g");
    group.setAttribute("id", `dot-${id}`);

    // Front circle
    const path1 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    path1.setAttribute("d", "M31.99 225.998c-5.718 3.301-5.67 8.681.106 12.016 5.777 3.336 15.096 3.364 20.815.063 5.718-3.302 5.67-8.682-.108-12.018-5.777-3.335-15.096-3.363-20.814-.061m5.26 3.037c2.842-1.64 7.474-1.627 10.346.03 2.873 1.658 2.896 4.333.054 5.974s-7.475 1.627-10.347-.03c-2.87-1.659-2.895-4.333-.053-5.974");
    path1.setAttribute("style", `fill:${light};`);
    path1.setAttribute("transform", `translate(${x - 0.988},${y - 196.054})`);

    // Back circle
    const path2 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    path2.setAttribute("d", "M32.203 222.032c-5.717 3.301-5.67 8.681.108 12.017 5.777 3.335 15.096 3.363 20.814.062s5.67-8.682-.107-12.018c-5.778-3.335-15.097-3.363-20.815-.061m5.261 3.037c2.843-1.64 7.475-1.627 10.347.03 2.872 1.659 2.896 4.333.053 5.974s-7.475 1.627-10.346-.03c-2.872-1.659-2.895-4.333-.054-5.974");
    path2.setAttribute("style", `fill:${dark};`);
    path2.setAttribute("transform", `translate(${x - 0.988},${y - 196.054})`);

    // Additional details
    const path3 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    path3.setAttribute("d", "M28.58 230.453a5.8 5.8 0 0 1-.625-2.507h-.036l-.177 3.579zM57.353 227.998c.007.828-.167 1.771-.573 2.563l.384 1.778z");
    path3.setAttribute("style", `fill:${light};`);
    path3.setAttribute("transform", `translate(${x - 0.988},${y - 196.054})`);

    group.appendChild(path1);
    group.appendChild(path2);
    group.appendChild(path3);

    return group;
}

function createLeftMerge(id: number, position: Position, colors: ColorPair): SVGElement {
    const { x, y } = position;
    const { light, dark } = colors;

    const group = document.createElementNS("http://www.w3.org/2000/svg", "g");
    group.setAttribute("id", `arrow-${id}`);

    const lightPath1 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    lightPath1.setAttribute("d", "m44.436 222.845-3.3 1.905-2.062-1.19-.824.476v.952l.824-.476 2.062 1.19 3.3-1.904z");
    lightPath1.setAttribute("style", `fill:${light}`);
    lightPath1.setAttribute("transform", `matrix(4.2 0 0 4.2 ${x - 131.204} ${y - 907.95})`);

    const lightPath2 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    lightPath2.setAttribute("d", "m41.549 222.131-.825.476v.953l.825-.477z");
    lightPath2.setAttribute("style", `fill:${light}`);
    lightPath2.setAttribute("transform", `matrix(4.2 0 0 4.2 ${x - 131.204} ${y - 907.95})`);

    const darkPath = document.createElementNS("http://www.w3.org/2000/svg", "path");
    darkPath.setAttribute("d", "m38.25 224.036.824-.476 2.062 1.19 3.3-1.905-1.65-.952-1.65.952-.412-.238.825-.476-2.887.238z");
    darkPath.setAttribute("style", `fill:${dark}`);
    darkPath.setAttribute("transform", `matrix(4.2 0 0 4.2 ${x - 131.204} ${y - 907.95})`);

    group.appendChild(darkPath);
    group.appendChild(lightPath1);
    group.appendChild(lightPath2);

    return group;
}

function createRightMerge(id: number, position: Position, colors: ColorPair): SVGElement {
    const { x, y } = position;
    const { light, dark } = colors;

    const group = document.createElementNS("http://www.w3.org/2000/svg", "g");
    group.setAttribute("id", `arrow-${id}`);

    const lightPath1 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    lightPath1.setAttribute("d", "m43.33 230.967 1.675-.967v-.967L43.33 230z");
    lightPath1.setAttribute("style", `fill:${light}`);
    lightPath1.setAttribute("transform", `matrix(4.13596 0 0 4.13596 ${x - 130.713} ${y - 919.273})`);

    const darkPath = document.createElementNS("http://www.w3.org/2000/svg", "path");
    darkPath.setAttribute("d", "m41.236 231.693.837-.484L39.98 230l3.35-1.934 1.675.967-1.675.967.419.242.851-.492-.433 1.701z");
    darkPath.setAttribute("style", `fill:${dark}`);
    darkPath.setAttribute("transform", `matrix(4.13596 0 0 4.13596 ${x - 130.713} ${y - 919.273})`);

    const lightPath2 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    lightPath2.setAttribute("d", "M39.98 230.967V230l2.093 1.21-.837.483z");
    lightPath2.setAttribute("style", `fill:${light}`);
    lightPath2.setAttribute("transform", `matrix(4.13596 0 0 4.13596 ${x - 130.713} ${y - 919.273})`);

    group.appendChild(darkPath);
    group.appendChild(lightPath1);
    group.appendChild(lightPath2);

    return group;
}

function createUpDown(id: number, position: Position, colors: ColorPair): SVGElement {
    const { x, y } = position;
    const { light, dark } = colors;

    const group = document.createElementNS("http://www.w3.org/2000/svg", "g");
    group.setAttribute("id", `arrow-${id}`);

    const lightPath = document.createElementNS("http://www.w3.org/2000/svg", "path");
    lightPath.setAttribute("d", "M55.898 222.727 50.229 226l-1.89-1.09V226l1.89 1.091 5.669-3.273z");
    lightPath.setAttribute("style", `fill:${light}`);
    lightPath.setAttribute("transform", `matrix(3.66673 0 0 3.66673 ${x - 149.536} ${y - 788.68})`);

    const darkPath = document.createElementNS("http://www.w3.org/2000/svg", "path");
    darkPath.setAttribute("d", "m54.008 221.637-5.668 3.272L50.23 226l5.668-3.273z");
    darkPath.setAttribute("style", `fill:${dark}`);
    darkPath.setAttribute("transform", `matrix(3.66673 0 0 3.66673 ${x - 149.536} ${y - 788.68})`);

    group.appendChild(darkPath);
    group.appendChild(lightPath);

    return group;
}

function createLeftRight(id: number, position: Position, colors: ColorPair): SVGElement {
    const { x, y } = position;
    const { light, dark } = colors;

    const group = document.createElementNS("http://www.w3.org/2000/svg", "g");
    group.setAttribute("id", `arrow-${id}`);

    const lightPath = document.createElementNS("http://www.w3.org/2000/svg", "path");
    lightPath.setAttribute("d", "M43.458 220.091v.97l5.04 2.909 1.679-.97v-.97l-1.68.97z");
    lightPath.setAttribute("style", `fill:${light}`);
    lightPath.setAttribute("transform", `matrix(4.12488 0 0 4.12488 ${x - 151.548} ${y - 879.85})`);

    const darkPath = document.createElementNS("http://www.w3.org/2000/svg", "path");
    darkPath.setAttribute("d", "m45.138 219.121-1.68.97 5.04 2.91 1.679-.97z");
    darkPath.setAttribute("style", `fill:${dark}`);
    darkPath.setAttribute("transform", `matrix(4.12488 0 0 4.12488 ${x - 151.548} ${y - 879.85})`);

    group.appendChild(darkPath);
    group.appendChild(lightPath);

    return group;
}

function createCrossing(id: number, position: Position, colors1: ColorPair, colors2: ColorPair): SVGElement {
    const { x, y } = position;
    const { light: light1, dark: dark1 } = colors1;
    const { light: light2, dark: dark2 } = colors2;

    const group = document.createElementNS("http://www.w3.org/2000/svg", "g");
    group.setAttribute("id", `crossing-${id}`);

    const lightPath1 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    lightPath1.setAttribute("d", "M55.426 28 34.64 40l-6.928-4v4l6.928 4 20.785-12Z");
    lightPath1.setAttribute("style", `fill:${light1}`);
    lightPath1.setAttribute("transform", `translate(${x}, ${y})`);

    const darkPath1 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    darkPath1.setAttribute("d", "M48.497 24 27.713 36l6.928 4 20.785-12Z");
    darkPath1.setAttribute("style", `fill:${dark1}`);
    darkPath1.setAttribute("transform", `translate(${x}, ${y})`);

    const darkPath2 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    darkPath2.setAttribute("d", "m34.641 24-6.928 4 5.196 3v-4l10.392 6v4l5.196 3 6.929-4-5.197-3v-4l-10.392-6-3.464 2z");
    darkPath2.setAttribute("style", `fill:${dark2}`);
    darkPath2.setAttribute("transform", `translate(${x}, ${y})`);

    const lightPath2 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    lightPath2.setAttribute("d", "M27.713 28v4l3.464 2 3.464-2 6.928 4v4l6.928 4 6.929-4v-4l-6.929 4-5.196-3v-4L32.91 27v4z");
    lightPath2.setAttribute("style", `fill:${light2}`);
    lightPath2.setAttribute("transform", `translate(${x}, ${y})`);

    group.appendChild(darkPath1);
    group.appendChild(lightPath1);
    group.appendChild(darkPath2);
    group.appendChild(lightPath2);

    return group;
}

function createLeftBranch(id: number, position: Position, colors: ColorPair): SVGElement {
    const { x, y } = position;
    const { light, dark } = colors;

    const group = document.createElementNS("http://www.w3.org/2000/svg", "g");
    group.setAttribute("id", `arrow-${id}`);

    const lightPath1 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    lightPath1.setAttribute("d", "m48.497 40-6.928-4-1.732 1v4l1.732-1 6.928 4z");
    lightPath1.setAttribute("style", `fill:${light}`);
    lightPath1.setAttribute("transform", `translate(${x}, ${y})`);

    const lightPath2 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    lightPath2.setAttribute("d", "m43.301 39-12.124-1-1.732-7v4l1.732 7L43.3 43zM55.426 36l-6.929 4v4l6.929-4Z");
    lightPath2.setAttribute("style", `fill:${light}`);
    lightPath2.setAttribute("transform", `translate(${x}, ${y})`);

    const darkPath = document.createElementNS("http://www.w3.org/2000/svg", "path");
    darkPath.setAttribute("d", "m29.445 31 3.464 2 8.66-5 13.857 8-6.929 4-6.928-4-1.732 1 3.464 2-12.124-1z");
    darkPath.setAttribute("style", `fill:${dark}`);
    darkPath.setAttribute("transform", `translate(${x}, ${y})`);

    group.appendChild(darkPath);
    group.appendChild(lightPath1);
    group.appendChild(lightPath2);

    return group;
}

function createRightBranch(id: number, position: Position, colors: ColorPair): SVGElement {
    const { x, y } = position;
    const { light, dark } = colors;

    const group = document.createElementNS("http://www.w3.org/2000/svg", "g");
    group.setAttribute("id", `arrow-${id}`);

    const lightPath1 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    lightPath1.setAttribute("d", "m48.497 32-6.928 4-1.732 1v4l1.732-1 6.928-4z");
    lightPath1.setAttribute("style", `fill:${light}`);
    lightPath1.setAttribute("transform", `translate(${x}, ${y})`);

    const lightPath2 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    lightPath2.setAttribute("d", "m43.301 39-12.124-1-1.732-7v4l1.732 7L43.3 43z");
    lightPath2.setAttribute("style", `fill:${light}`);
    lightPath2.setAttribute("transform", `translate(${x}, ${y})`);

    const lightPath3 = document.createElementNS("http://www.w3.org/2000/svg", "path");
    lightPath3.setAttribute("d", "m34.641 32-6.928-4v4l6.928 4z");
    lightPath3.setAttribute("style", `fill:${light}`);
    lightPath3.setAttribute("transform", `translate(${x}, ${y})`);

    const darkPath = document.createElementNS("http://www.w3.org/2000/svg", "path");
    darkPath.setAttribute("d", "m29.445 31 3.464 2 1.732-1-6.928-4 6.928-4 13.856 8-8.66 5 3.464 2-12.124-1z");
    darkPath.setAttribute("style", `fill:${dark}`);
    darkPath.setAttribute("transform", `translate(${x}, ${y})`);

    group.appendChild(darkPath);
    group.appendChild(lightPath1);
    group.appendChild(lightPath2);
    group.appendChild(lightPath3);

    return group;
}

function createHorizontalLines(nextRowState: (RowState | null)[], start: number, end: number, color: ColorPair, nr: number) {
    const step = start < end ? 1 : -1;
    const distance = Math.abs(end - start);

    for (let j = 1; j < distance; j++) {
        const pos = start + (j * step);
        const position = calculatePosition(pos, nr);
        if (nextRowState[pos] && nextRowState[pos]?.type === 'line') {
            const cross = createCrossing(Date.now(), position, nextRowState[pos]!.color, color);
            nextRowState[pos] = {
                type: 'cross',
                color: nextRowState[pos]!.color,
                branch: start,
                element: cross,
                index: pos,
                nr: nr
            };
        } else {
            const hline = createLeftRight(Date.now(), position, color);
            nextRowState[pos] = {
                type: 'hline',
                color: color,
                branch: start,
                element: hline,
                index: pos,
                nr: nr
            };
        }
    }
}

function getRandomColorPair(rowState: (RowState | null)[]): ColorPair {
    const usedColors = rowState
        .filter((state): state is RowState => state !== null)
        .map(state => state.color);

    const availableColors = Object.values(COLORS).filter(color =>
        !usedColors.some(usedColor =>
            usedColor.light === color.light && usedColor.dark === color.dark
        )
    );

    if (availableColors.length === 0) {
        return Object.values(COLORS)[Math.floor(Math.random() * Object.values(COLORS).length)];
    }

    return availableColors[Math.floor(Math.random() * availableColors.length)];
}

function calculatePosition(xCoord: number, yCoord: number, tileOffset = { x: 3, y: -19 }): Position {
    const tileWidth = 24;
    const tileHeight = 12;
    const isoXOffset = tileWidth * 0.866025404;

    const offsetX = xCoord - tileOffset.x;
    const offsetY = yCoord - tileOffset.y;

    const baseX = 0;
    const baseY = 0;

    const isoX = (offsetX - offsetY) * isoXOffset;
    const isoY = (offsetX + offsetY) * tileHeight;

    return {
        x: baseX + isoX + isoXOffset,
        y: baseY + isoY + tileHeight
    };
}

function createTile(x: number, y: number, isEven: boolean): SVGGElement {
    const tileWidth: number = 24;
    const tileHeight: number = 12;
    const isoXOffset: number = (tileWidth * 0.866025404);  // cos(30°) ≈ 0.866
    const depth: number = 3;

    let baseX: number = (x - y) * isoXOffset;
    let baseY: number = (x + y) * tileHeight;
    const group: SVGGElement = document.createElementNS("http://www.w3.org/2000/svg", "g");

    const colors: TileColors = isEven ? {
        top: "#ddd0c8",
        front: "#eaebdb",
        right: "#9f958b"
    } : {
        top: "#ece5df",
        front: "#eaebdb",
        right: "#9f958b"
    };

    // Top face
    const topFace: SVGPathElement = document.createElementNS("http://www.w3.org/2000/svg", "path");
    topFace.setAttribute("d", `m ${baseX},${baseY} l ${isoXOffset},${tileHeight} ${isoXOffset},-${tileHeight} -${isoXOffset},-${tileHeight} z`);
    topFace.setAttribute("fill", colors.top);
    group.appendChild(topFace);

    // Front face
    const frontFace: SVGPathElement = document.createElementNS("http://www.w3.org/2000/svg", "path");
    frontFace.setAttribute("d", `m ${baseX},${baseY} l ${isoXOffset},${tileHeight} 0,${depth} -${isoXOffset},-${tileHeight} z`);
    frontFace.setAttribute("fill", colors.front);
    group.appendChild(frontFace);

    // Right face
    const rightFace: SVGPathElement = document.createElementNS("http://www.w3.org/2000/svg", "path");
    rightFace.setAttribute("d", `m ${baseX + isoXOffset},${baseY + tileHeight} l ${isoXOffset},-${tileHeight} 0,${depth} -${isoXOffset},${tileHeight} z`);
    rightFace.setAttribute("fill", colors.right);
    group.appendChild(rightFace);

    return group;
}

function createRow(y: number, width: number): SVGGElement {
    const row: SVGGElement = document.createElementNS("http://www.w3.org/2000/svg", "g");
    row.setAttribute("data-row", y.toString());
    for (let x = 0; x < width; x++) {
        const isEven: boolean = (x + y) % 2 === 0;
        const tile = createTile(x, y, isEven);
        row.appendChild(tile);
    }
    return row;
}

function buildFloor(width: number, height: number, config: FloorConfig): SVGElement {
    const group: SVGGElement = document.createElementNS("http://www.w3.org/2000/svg", "g");
    group.setAttribute("id", "floorGroup");
    //group.setAttribute("transform", `translate(${config.viewOffsetX},${config.viewOffsetY})`);

    for (let y = 0; y < height; y++) {
        const row = createRow(y, width);
        group.appendChild(row);
    }

    return group
}

function removeTopRow(group: HTMLElement) {
    if (group.firstChild) {
        group.removeChild(group.firstChild);
    }
}

function addRowToBottom(group: HTMLElement, width: number): SVGElement {
    const lastRow = group.lastChild as SVGElement;
    const lastRowIndex = parseInt(lastRow?.getAttribute("data-row") || "-1");
    const newRow = createRow(lastRowIndex + 1, width);
    group.appendChild(newRow);
    return newRow;
}

function updateMovement(deltaTime: number): boolean {
    const diff = MOVEMENT_SPEED * (deltaTime / 1000);

    movement.totalDistance += diff;
    movement.currentTileProgress += diff;

    if (movement.currentTileProgress >= movement.tileHeight) {
        movement.currentTileProgress -= movement.tileHeight;
        return true;
    }
    return false;
}

let nr = 1;
let currentRowState: (RowState | null)[] = Array(12).fill(null);
let dots: SVGElement[] = [];
let lastFrameTime = 0;

export function initAnimation( ): () => void {
    function animate(timestamp?: number) {
        if (!timestamp) {
            requestAnimationFrame(animate);
            return;
        }
        //console.log("timestamp", timestamp, lastFrameTime);
        //console.log("called", timestamp);

        const deltaTime = timestamp - lastFrameTime;


        if (updateMovement(deltaTime) ) {


            const group = document.getElementById('floorGroup');
            if (!group) return;

            removeTopRow(group);
            for (let i = 0; i < 12; i++) {
                if (currentRowState[i] && currentRowState[i]!.nr > 20) {
                    const dotsGroup = document.getElementById('dots');
                    if (dotsGroup?.firstChild) {
                        dotsGroup.removeChild(dotsGroup.firstChild);
                    }
                    break;
                }
            }

            addRowToBottom(group, 12);
            const nextRowState: (RowState | null)[] = Array(12).fill(null);

            // First row
            if (nr === 1) {
                for (let i = 0; i < 3; i++) {
                    const randomColor = getRandomColorPair(nextRowState);
                    const position = calculatePosition(i * 4, nr);
                    const newDot = createDot(Date.now(), position, randomColor);
                    nextRowState[i * 4] = {
                        type: 'dot',
                        color: randomColor,
                        branch: i * 4,
                        element: newDot,
                        index: i * 4,
                        nr: nr
                    };
                }


                // Branch logic for first row
                for (let i = 0; i < 12; i++) {
                    if (nextRowState[i]?.type === 'dot') {
                        if (Math.random() < 0.5) {
                            // Check both directions
                            // Check left direction - scan until we hit a dot or reach the end
                            let leftTargets = -1;
                            for (let left = i + 1; left < 12; left++) {
                                if (!nextRowState[left]) {
                                    leftTargets = left;
                                    break;
                                } else if (nextRowState[left]?.type === 'dot') {
                                    break
                                }
                            }


                            // Check right direction - scan until we hit a dot or reach the start
                            let rightTargets = -1;
                            for (let right = i - 1; right >= 0; right--) {
                                if (!nextRowState[right]) {
                                    rightTargets = right;
                                    break;
                                } else if (nextRowState[right]?.type === 'dot') {
                                    break;
                                }
                            }


                            // If we have possible directions, randomly choose one
                            if (leftTargets >= 0 && rightTargets >= 0) {
                                if (Math.random() < 0.5) {
                                    // Choose random left target
                                    const targetIndex = leftTargets;
                                    const position = calculatePosition(targetIndex, nr);
                                    const randomColor = getRandomColorPair(nextRowState);
                                    const leftArrow = createRightBranch(Date.now(), position, randomColor);
                                    nextRowState[targetIndex] = {
                                        type: 'branch-left',
                                        color: randomColor,
                                        branch: i,
                                        element: leftArrow,
                                        index: targetIndex,
                                        nr: nr
                                    };
                                } else {
                                    // Choose random right target
                                    const targetIndex = rightTargets;
                                    const position = calculatePosition(targetIndex, nr);
                                    const randomColor = getRandomColorPair(nextRowState);
                                    const rightArrow = createLeftBranch(Date.now(), position, randomColor);
                                    nextRowState[targetIndex] = {
                                        type: 'branch-right',
                                        color: randomColor,
                                        branch: i,
                                        element: rightArrow,
                                        index: targetIndex,
                                        nr: nr
                                    };
                                }
                            } else if (leftTargets >= 0) {
                                // Choose random left target
                                const targetIndex = leftTargets;
                                const position = calculatePosition(targetIndex, nr);
                                const randomColor = getRandomColorPair(nextRowState);
                                const leftArrow = createRightBranch(Date.now(), position, randomColor);
                                nextRowState[targetIndex] = {
                                    type: 'branch-left',
                                    color: randomColor,
                                    branch: i,
                                    element: leftArrow,
                                    index: targetIndex,
                                    nr: nr
                                };
                            } else if (rightTargets >= 0) {
                                // Choose random right target
                                const targetIndex = rightTargets;
                                const position = calculatePosition(targetIndex, nr);
                                const randomColor = getRandomColorPair(nextRowState);
                                const rightArrow = createLeftBranch(Date.now(), position, randomColor);
                                nextRowState[targetIndex] = {
                                    type: 'branch-right',
                                    color: randomColor,
                                    branch: i,
                                    element: rightArrow,
                                    index: targetIndex,
                                    nr: nr
                                };
                            }
                        }
                    }}
            } else {
                //console.log("called", nr)
                // All other rows
                for (let i = 0; i < 12; i++) {
                    if (currentRowState[i]) {
                        if (currentRowState[i]!.type === 'dot' ||
                            currentRowState[i]!.type === 'branch-right' ||
                            currentRowState[i]!.type === 'branch-left' ||
                            currentRowState[i]!.type === 'cross' ||
                            currentRowState[i]!.type === 'line') {
                            const r = Math.random();
                            if (r < 0.2 && currentRowState[i]!.type !== 'dot') {
                                //commit
                                const position = calculatePosition(i, nr);
                                const dot = createDot(Date.now(), position, currentRowState[i]!.color);
                                nextRowState[i] = {
                                    type: 'dot',
                                    color: currentRowState[i]!.color,
                                    branch: currentRowState[i]!.branch,
                                    element: dot,
                                    index: i,
                                    nr: nr
                                };
                                //console.log("put dot since previsous dot was ", currentRowState[i]!.color, i, nr)
                            } else {
                                //continue
                                const position = calculatePosition(i, nr);
                                const line = createUpDown(Date.now(), position, currentRowState[i]!.color);
                                nextRowState[i] = {
                                    type: 'line',
                                    color: currentRowState[i]!.color,
                                    branch: currentRowState[i]!.branch,
                                    element: line,
                                    index: i,
                                    nr: nr
                                };
                                //console.log("put line since previsous dot was ", currentRowState[i]!.color, i, nr)
                            }
                        }
                    }
                }

                // Merge logic
                for (let i = 0; i < 12; i++) {
                    if (currentRowState[i] && nextRowState[i] && nextRowState[i]!.type === 'line') {
                        if (currentRowState[i]!.type === 'dot' ||
                            currentRowState[i]!.type === 'branch-right' ||
                            currentRowState[i]!.type === 'branch-left' ||
                            currentRowState[i]!.type === 'cross' ||
                            currentRowState[i]!.type === 'line') {

                            const r = Math.random();
                            if (r < 0.5 && currentRowState[i]!.branch !== i) {
                                //merge here
                                let origin = nextRowState[i]!.branch;
                                const isGoingLeft = origin < i;
                                const start = Math.min(i, origin);
                                const end = Math.max(i, origin);

                                let canReachOrigin = nextRowState[origin] && nextRowState[origin].type === 'dot';
                                for (let j = start + 1; j < end; j++) {
                                    if (nextRowState[j]?.type === 'dot' || nextRowState[j]?.type === 'cross' || nextRowState[j]?.type === 'merge-left' || nextRowState[j]?.type === 'merge-right') {
                                        canReachOrigin = false;
                                        break;
                                    }
                                }

                                if (canReachOrigin) {
                                    // Create merge arrow based on direction
                                    const position = calculatePosition(i, nr);
                                    if (isGoingLeft) {
                                        const leftArrow = createLeftMerge(Date.now(), position, currentRowState[i]!.color);
                                        //remove the previous
                                        nextRowState[i] = {
                                            type: 'merge-left',
                                            color: currentRowState[i]!.color,
                                            branch: origin,
                                            element: leftArrow,
                                            index: i,
                                            nr: nr
                                        };

                                    } else {
                                        const rightArrow = createRightMerge(Date.now(), position, currentRowState[i]!.color);
                                        //remove the previous
                                        nextRowState[i] = {
                                            type: 'merge-right',
                                            color: currentRowState[i]!.color,
                                            branch: origin,
                                            element: rightArrow,
                                            index: i,
                                            nr: nr
                                        };

                                    }

                                    // Create horizontal lines connecting to origin
                                    if (Math.abs(origin - i) > 1) {
                                        createHorizontalLines(nextRowState, i, origin, currentRowState[i]!.color, nr);
                                    }
                                }
                            }
                        }
                    }
                }

                // Branch logic
                for (let i = 0; i < 12; i++) {
                    if (nextRowState[i] && nextRowState[i]?.type === 'dot') {
                        if (Math.random() < 0.5) {
                            // Check both directions
                            // Check left direction - scan until we hit a dot or reach the end
                            let leftTargets = -1;
                            let c = 0;
                            for (let left = i + 1; left < 12; left++) {
                                c++
                                if (!nextRowState[left]) {
                                    leftTargets = left;
                                    break;
                                } else if (nextRowState[left]?.type === 'dot' || nextRowState[left]?.type === 'merge-left' || nextRowState[left]?.type === 'merge-right' || nextRowState[left]?.type === 'hline') {
                                    break
                                } else if (c > 1) {
                                    break
                                }
                            }

                            // Check right direction - scan until we hit a dot or reach the start
                            let rightTargets = -1;
                            c = 0;
                            for (let right = i - 1; right >= 0; right--) {
                                c++
                                if (!nextRowState[right]) {
                                    rightTargets = right;
                                    break;
                                } else if (nextRowState[right]?.type === 'dot' || nextRowState[right]?.type === 'merge-left' || nextRowState[right]?.type === 'merge-right' || nextRowState[right]?.type === 'hline') {
                                    break;
                                } else if (c > 1) {
                                    break
                                }
                            }

                            // If we have possible directions, randomly choose one
                            if (leftTargets >= 0 && rightTargets >= 0) {
                                if (Math.random() < 0.5) {
                                    // Choose random left target
                                    const targetIndex = leftTargets;
                                    const randomColor = getRandomColorPair(nextRowState);

                                    // Create horizontal lines for gap
                                    if (targetIndex - i !== 0) {
                                        createHorizontalLines(nextRowState, i, targetIndex, randomColor, nr);
                                    }

                                    const position = calculatePosition(targetIndex, nr);
                                    const leftArrow = createRightBranch(Date.now(), position, randomColor);
                                    nextRowState[targetIndex] = {
                                        type: 'branch-left',
                                        color: randomColor,
                                        branch: currentRowState[i]!.branch,
                                        element: leftArrow,
                                        index: targetIndex,
                                        nr: nr
                                    };
                                } else {
                                    // Choose random right target
                                    const targetIndex = rightTargets;
                                    const position = calculatePosition(targetIndex, nr);
                                    const randomColor = getRandomColorPair(nextRowState);

                                    const rightArrow = createLeftBranch(Date.now(), position, randomColor);
                                    nextRowState[targetIndex] = {
                                        type: 'branch-right',
                                        color: randomColor,
                                        branch: currentRowState[i]!.branch,
                                        element: rightArrow,
                                        index: targetIndex,
                                        nr: nr
                                    };

                                    // Create horizontal lines for gap
                                    if (targetIndex - i !== 0) {
                                        createHorizontalLines(nextRowState, i, targetIndex, randomColor, nr);
                                    }
                                }
                            } else if (leftTargets >= 0) {
                                // Choose random left target
                                const targetIndex = leftTargets;
                                const position = calculatePosition(targetIndex, nr);
                                const randomColor = getRandomColorPair(nextRowState);

                                // Create horizontal lines for gap
                                if (targetIndex - i !== 0) {
                                    createHorizontalLines(nextRowState, i, targetIndex, randomColor, nr);
                                }

                                const leftArrow = createRightBranch(Date.now(), position, randomColor);
                                nextRowState[targetIndex] = {
                                    type: 'branch-left',
                                    color: randomColor,
                                    branch: currentRowState[i]!.branch,
                                    element: leftArrow,
                                    index: targetIndex,
                                    nr: nr
                                };
                            } else if (rightTargets >= 0) {
                                // Choose random right target
                                const targetIndex = rightTargets;
                                const position = calculatePosition(targetIndex, nr);
                                const randomColor = getRandomColorPair(nextRowState);

                                const rightArrow = createLeftBranch(Date.now(), position, randomColor);
                                nextRowState[targetIndex] = {
                                    type: 'branch-right',
                                    color: randomColor,
                                    branch: i,
                                    element: rightArrow,
                                    index: targetIndex,
                                    nr: nr
                                };

                                // Create horizontal lines for gap
                                if (targetIndex - i !== 0) {
                                    createHorizontalLines(nextRowState, i, targetIndex, randomColor, nr);
                                }
                            }
                        }
                    }
                }
            }

            currentRowState = [...nextRowState];
            // Sort and append new elements
            nextRowState.sort((a, b) => {
                if (!a) return 1;
                if (!b) return -1;
                return a.index - b.index;
            });

            const groupX = document.createElementNS("http://www.w3.org/2000/svg", "g");
            for (const state of nextRowState) {
                if (state) {
                    groupX.appendChild(state.element);
                }
            }

            const dotsGroup = document.getElementById('dots');
            if (dotsGroup) {
                dotsGroup.appendChild(groupX);
            }
            dots.push(groupX);

            nr++;
        }

        // Update positions
        const xOffset = movement.totalDistance * Math.sqrt(3);
        const xt = floorConfig.viewOffsetX + xOffset;
        const yt = floorConfig.viewOffsetY - movement.totalDistance;

        dots = dots.filter(dot => {
            if (!dot.isConnected) {
                return false;
            }
            dot.setAttribute("transform", `translate(${xt},${yt})`);
            return true;
        });

        lastFrameTime = timestamp;
        requestAnimationFrame(animate);
    }

    const element = buildFloor(12, 24, floorConfig);
    const svg = document.getElementById('scrollingFloor');
    if(svg) {
        svg.appendChild(element);
    }
    dots.push(element);

    requestAnimationFrame(animate);
    // Return cleanup function
    return () => {

    };
}