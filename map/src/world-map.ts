import {
    Application,
    Container,
    Graphics,
    MIPMAP_MODES,
    MSAA_QUALITY,
    Point,
    SCALE_MODES,
    Sprite,
    TARGETS,
    Texture
} from "pixi.js";
import {Viewport} from "pixi-viewport";
import {Stats} from "stats.ts";
import {Vec2, Vec3} from "./util";
import {Marker} from "./workers/events";
import {quadtree, Quadtree} from "d3-quadtree";

export type TileUrlFunc = (coord: Vec2, zoom: number) => string;

export type MarkerClickFunc = (marker: Marker, global: Point, local: Point) => void;

export interface WorldMapConfig {
    stats: Stats | undefined,
    tileSize: Vec2,
    bounds: Vec2,
    minNativeZoom: number,
    maxNativeZoom: number,
    getTileUrl: TileUrlFunc,
    onMarkerClick: MarkerClickFunc,
    crs: {
        rotation: number
        offset: Vec2,
        gridSize: Vec2,
        scale: Vec2,
    },
}

export class WorldMap {
    readonly viewport: Viewport;
    readonly tileContainer: Container = new Container();

    readonly markerContainer: Container = new Container();
    readonly polygonGfx: Graphics = new Graphics();
    readonly markerGfx: Graphics = new Graphics();
    private markerPoints: Quadtree<Marker>;

    private lastNativeZoom: number;
    private spriteCache = new Map<Vec3, Sprite>();

    constructor(
        readonly app: Application,
        readonly config: WorldMapConfig,
    ) {
        // Start viewport setup
        this.viewport = new Viewport({
            screenWidth: app.view.width,
            screenHeight: app.view.height,
            worldWidth: this.config.tileSize.x * this.config.bounds.x,
            worldHeight: this.config.tileSize.y * this.config.bounds.y,
            events: app.renderer.events,
            ticker: app.ticker,
        });
        this.viewport.sortableChildren = true;

        // Setup container z indexes
        this.tileContainer.zIndex = 0;
        this.markerContainer.zIndex = 1;
        this.polygonGfx.zIndex = 2;
        this.markerGfx.zIndex = 3;

        // Setup tile container
        this.tileContainer.interactive = true;
        this.viewport.addChild(this.tileContainer);

        // Setup markers
        this.markerPoints = quadtree<Marker>()
            .x((m) => m.x)
            .y((m) => m.y);

        this.markerContainer.addChild(this.markerGfx);
        this.markerContainer.addChild(this.polygonGfx);
        this.markerContainer.interactive = true;
        this.markerContainer.on('click', (e) => {
            const local = this.markerContainer.toLocal(e.global);
            const marker = this.markerPoints.find(local.x, local.y, 5);
            if (!marker) {
                return
            }
            this.config.onMarkerClick(marker, e.global, local);
        });

        this.viewport.addChild(this.markerContainer);

        // Finish viewport setup
        this.app.stage.interactive = true;
        this.app.stage.addChild(this.viewport);

        this.viewport
            .drag()
            .pinch()
            .wheel();

        this.lastNativeZoom = this.nativeZoom;

        // Start rendering
        app.ticker.maxFPS = 60;
        app.ticker.add(() => {
            this.config.stats?.begin();

            if (this.viewport.dirty) {
                // Zoom change
                const nextNativeZoom = this.nativeZoom;
                if (this.lastNativeZoom != nextNativeZoom) {
                    this.onNativeZoomChange(this.lastNativeZoom, nextNativeZoom);
                    this.lastNativeZoom = nextNativeZoom;
                }

                // Viewport update
                this.draw();
                this.drawMarkers();
                this.viewport.dirty = false;
            }

            this.config.stats?.end();
        });

        // Draw center of world
        this.markerGfx.beginFill(0x000000, 1);
        this.markerGfx.drawCircle(0, 0, 0.25);
        this.markerGfx.endFill();
    }

    public resize(width: number, height: number) {
        this.app.renderer.resize(width, height);
        this.viewport.resize(width, height);
    }

    get nativeZoom(): number {
        let l = Math.ceil((Math.log2(this.viewport.scaled)));
        return Math.min(Math.max(l, this.config.minNativeZoom), this.config.maxNativeZoom);
    }

    getTileTexture(tileCoord: Vec2): Texture | undefined {
        const tileUrl = this.config.getTileUrl(tileCoord, this.nativeZoom);
        if (tileUrl == '') {
            return undefined;
        }
        return Texture.from(tileUrl, {
            mipmap: MIPMAP_MODES.ON,
            anisotropicLevel: 2,
            scaleMode: SCALE_MODES.LINEAR,
            width: this.config.tileSize.x,
            height: this.config.tileSize.y,
            target: TARGETS.TEXTURE_2D,
            multisample: MSAA_QUALITY.HIGH,
        });
    }

    getCurrentScale(): number {
        return 1 / Math.pow(2, this.config.maxNativeZoom - this.nativeZoom);
    }

    getCurrentBounds(): Vec2 {
        return this.config.bounds.scale(this.getCurrentScale());
    }

    getTilePos(tileCoord: Vec2): Vec2 {
        return tileCoord.mul(this.config.tileSize);
    }

    getTileSprite(tileCoord: Vec2): Sprite | undefined {
        const tilePos = this.getTilePos(tileCoord);

        const tex = this.getTileTexture(tileCoord);
        if (!tex) {
            return undefined;
        }

        const sprite = new Sprite(tex);
        sprite.x = tilePos.x;
        sprite.y = tilePos.y;
        return sprite
    }

    eachTileInView(cb: (tileCoord: Vec2) => void) {
        const visBounds = this.viewport.getVisibleBounds();
        const bounds = this.getCurrentBounds();

        let xMin = Math.floor(visBounds.x / this.config.tileSize.x);
        let yMin = Math.floor(visBounds.y / this.config.tileSize.y);
        let xMax = Math.ceil((visBounds.x + visBounds.width) / this.config.tileSize.x);
        let yMax = Math.ceil((visBounds.y + visBounds.height) / this.config.tileSize.y);

        if (xMin < 0) {
            xMin = 0;
        }
        if (yMin < 0) {
            yMin = 0;
        }
        if (xMax > bounds.x) {
            xMax = bounds.x;
        }
        if (yMax > bounds.y) {
            yMax = bounds.y;
        }

        for (let x = xMin; x < xMax; x++) {
            for (let y = yMin; y < yMax; y++) {
                cb(new Vec2(x, y));
            }
        }
    }

    onNativeZoomChange(lastZoom: number, newZoom: number) {
        const ratio = Math.pow(2, newZoom) / Math.pow(2, lastZoom);
        this.viewport.center = new Point(this.viewport.center.x * ratio, this.viewport.center.y * ratio);
    }

    drawMarkers() {
        const scale = this.getCurrentScale()
            * (this.config.tileSize.x / this.config.crs.gridSize.x)
            / this.config.crs.scale.x;

        this.markerContainer.setTransform(
            this.config.crs.offset.x * scale,
            this.config.crs.offset.y * scale,
            this.config.crs.scale.x * scale,
            this.config.crs.scale.y * scale,
            this.config.crs.rotation,
        );
    }

    draw() {
        // Clear viewport
        this.tileContainer.removeChildren(0);

        // Draw tiles
        this.eachTileInView((tileCoord: Vec2) => {
            const tileCoordWithZoom = tileCoord.withZ(tileCoord, this.nativeZoom);

            let sprite: Sprite | undefined = this.spriteCache.get(tileCoordWithZoom);
            if (!sprite) {
                // Create sprite
                sprite = this.getTileSprite(tileCoord);
                if (sprite) {
                    this.spriteCache.set(tileCoordWithZoom, sprite);
                    this.tileContainer.addChild(sprite);
                }
            }
        });
    }

    addMarker(m: Marker) {
        this.markerGfx.beginFill(m.color, 0.5);
        // this.markerGfx.drawRect(m.x, m.y, m.width, m.height);
        this.markerGfx.drawCircle(m.x, m.y, m.w / 2);
        this.markerGfx.endFill();
        this.markerPoints.add(m);
    }

    addPolygon(p: Array<Point>) {
        this.polygonGfx.lineStyle({
            width: 2,
            color: 0xACA491,
            alpha: 0.5,
        });
        this.polygonGfx.drawPolygon(p);
        this.polygonGfx.lineStyle();
    }

    clearMarkers() {
        this.markerGfx.clear();
        this.markerPoints = quadtree<Marker>()
            .x(this.markerPoints.x())
            .y(this.markerPoints.y());
    }
}