import {Buffer, Geometry, Mesh, Shader, utils} from "pixi.js";
import {TYPES} from "@pixi/core";

const shader = Shader.from(`precision mediump float;

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


    constructor(
        readonly maxSize: number
    ) {
        this.data = new Float32Array(maxSize * (this.positionSize + this.colorSize));
    }

    public clear() {
        this.data = new Float32Array(this.data.length);
        this.size = 0;
        this.count = 0;
    }

    public addMarker(x: number, y: number, color: number) {
        const rgb = utils.hex2rgb(color);

        this.data[this.size] = x;
        this.data[this.size + 1] = y;
        this.data[this.size + 2] = rgb[0];
        this.data[this.size + 3] = rgb[1];
        this.data[this.size + 4] = rgb[2];

        this.size += this.positionSize + this.colorSize;
        this.count += 1;
    }

    public getMesh(markerSize: number): Mesh<Shader> {
        const buffer = new Buffer(this.data.slice(0, this.size));

        const geometry = new Geometry()
            .addAttribute(
                'aVPos',
                [
                    -markerSize, -markerSize,
                    markerSize, -markerSize,
                    markerSize, markerSize,
                    -markerSize, markerSize
                ],
                2,
            )
            .addAttribute(
                'aIPos',
                buffer,
                this.positionSize,
                false,
                TYPES.FLOAT,
                4 * (this.positionSize + this.colorSize),
                0,
                true,
            )
            .addAttribute(
                'aICol',
                buffer,
                this.colorSize,
                false,
                TYPES.FLOAT,
                4 * (this.positionSize + this.colorSize),
                4 * this.positionSize,
                true,
            )
            .addIndex([0, 1, 2, 0, 2, 3]);

        geometry.instanced = true;
        geometry.instanceCount = this.count;

        return new Mesh(geometry, shader);
    }
}