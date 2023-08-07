import {Buffer, Geometry, Mesh, Shader, utils} from "pixi.js";
import {TYPES, DRAW_MODES} from "@pixi/core";

const DoublePi = 2.0 * Math.PI;
const MarkerShader = Shader.from(`precision mediump float;

    attribute vec2 aVPos;
    attribute vec2 aIPos;
    attribute vec3 aICol;

    uniform mat3 translationMatrix;
    uniform mat3 projectionMatrix;

    varying vec3 vCol;

    void main() {
        vCol = aICol;
        gl_Position = vec4((projectionMatrix * translationMatrix * vec3(aVPos + aIPos, 1.0)).xy, 0.0, 1.0);
    }`,

    `precision mediump float;

varying vec3 vCol;

void main() {
    gl_FragColor = vec4(vCol, 0.50);
}`);

export class MarkerMesh {
    readonly positionSize = 2;
    readonly colorSize = 3;

    private data: Float32Array;
    private size: number = 0;
    private count: number = 0;

    private mesh: Mesh<Shader> | undefined;
    private _radius: number = 1;


    constructor(
        readonly maxSize: number,
        readonly markerSegments: number = 20,
    ) {
        this.data = new Float32Array(maxSize * (this.positionSize + this.colorSize));
    }

    get radius(): number {
        return this._radius;
    }

    set radius(r: number) {
        if (r == this._radius) {
            return;
        }

        this._radius = r;

        if (this.mesh) {
            const buf = this.mesh.geometry.getBuffer('aVPos');
            buf.update(this.generateVertices(r, this.markerSegments));
        }
    }

    public clear() {
        this.data = new Float32Array(this.data.length);
        this.mesh = undefined;
        this.size = 0;
        this.count = 0;
    }

    public addMarker(x: number, y: number, color: number) {
        this.mesh = undefined;

        const rgb = utils.hex2rgb(color);
        this.data[this.size++] = x;
        this.data[this.size++] = y;
        this.data[this.size++] = rgb[0];
        this.data[this.size++] = rgb[1];
        this.data[this.size++] = rgb[2];
        this.count++;
    }

    public update() {
        const data = this.data.slice(0, this.size);

        const geometry = new Geometry()
            .addAttribute(
                'aVPos',
                this.generateVertices(this._radius, this.markerSegments),
                2,
            )
            .addAttribute(
                'aIPos',
                data,
                this.positionSize,
                false,
                TYPES.FLOAT,
                4 * (this.positionSize + this.colorSize),
                0,
                true,
            )
            .addAttribute(
                'aICol',
                data,
                this.colorSize,
                false,
                TYPES.FLOAT,
                4 * (this.positionSize + this.colorSize),
                4 * this.positionSize,
                true,
            );
        geometry.instanced = true;
        geometry.instanceCount = this.count;

        this.mesh = new Mesh(geometry, MarkerShader, undefined, DRAW_MODES.TRIANGLE_FAN);
    }

    private generateVertices(scale: number, segments: number): Float32Array {
        const vertices = new Float32Array(segments * 2);
        let j = 0;
        for (let i = 0; i < segments; i++) {
            const theta = (i / segments) * DoublePi;
            vertices[j++] = Math.cos(theta) * scale;
            vertices[j++] = Math.sin(theta) * scale;
        }
        return vertices;
    }

    public getMesh(): Mesh<Shader> {
        if (!this.mesh) {
            this.update();
        }
        return this.mesh!;
    }
}