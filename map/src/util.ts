import Quadtree from "@timohausmann/quadtree-js";

export class Marker {
    constructor(
        readonly pos: Vec2,
        // TODO: color, etc
    ) {}
}

export class Vec2 {
    constructor(
        readonly x: number,
        readonly y: number,
    ) {}

    scale(scale: number): Vec2 {
        return new Vec2(this.x * scale, this.y * scale);
    }

    scaleGrid(gridSize: number, newGridSize: number): Vec2 {
        return this.scale(newGridSize / gridSize);
    }

    mul(scale: Vec2): Vec2 {
        return new Vec2(this.x * scale.x, this.y * scale.y);
    }

    div(scale: Vec2): Vec2 {
        return new Vec2(this.x / scale.x, this.y / scale.y);
    }

    add(v: Vec2): Vec2 {
        return new Vec2(this.x + v.x, this.y + v.y);
    }

    sub(v: Vec2): Vec2 {
        return new Vec2(this.x - v.x, this.y - v.y);
    }

    withZ(v: Vec2, z: number): Vec3 {
        return new Vec3(v.x, v.y, z);
    }

    rotate(angle: number): Vec2 {
        const rads = (Math.PI / 180) * angle;
        const cos = Math.cos(rads);
        const sin = Math.sin(rads);
        return new Vec2(
            (cos * this.x) + (sin * this.y),
            (cos * this.y) - (sin * this.x)
        );
    }
}

export class Vec3 {
    constructor(
        readonly x: number,
        readonly y: number,
        readonly z: number,
    ) {}
}

export function ClosestPoint<A extends {x: number, y: number}, B extends Quadtree.Rect>(target: A, rects: B[]): B | undefined {
    let closestDist = Number.POSITIVE_INFINITY;
    let result: B | undefined  = undefined;

    for (let rect of rects) {
        const pointX = rect.x + (rect.width / 2);
        const pointY = rect.y + (rect.height / 2);
        const dist = Math.hypot(pointX-target.x, pointY-target.y);
        if (dist < closestDist) {
            closestDist = dist;
            result = rect;
        }
    }

    return result
}
