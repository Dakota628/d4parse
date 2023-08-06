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
import $ from "jquery";
import {LRUCache} from "typescript-lru-cache";
import {MarkerMesh} from "./marker-mesh";

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
    coordinateDisplay: JQuery<HTMLElement> | undefined,
    crs: {
        rotation: number
        offset: Vec2,
        gridSize: Vec2,
        scale: Vec2,
    },
}

export class WorldMap {
    private nativeZoom: number = 0;
    private markerSize: number = 0;

    readonly viewport: Viewport;
    readonly tileContainer: Container = new Container();

    readonly markerContainer: Container = new Container();
    readonly polygonGfx: Graphics = new Graphics();
    readonly markerMesh = new MarkerMesh(1000000);

    private markerPoints: Quadtree<Marker>;
    private currentMarker: Marker | undefined = undefined;

    private spriteCache = new LRUCache<Vec3, Sprite>({
        maxSize: 1600,
    })

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
        this.polygonGfx.zIndex = 1;
        this.markerContainer.zIndex = 2;

        // Setup tile container
        this.tileContainer.interactive = true;
        this.viewport.addChild(this.tileContainer);

        // Setup markers
        this.markerPoints = quadtree<Marker>()
            .x((m) => m.x)
            .y((m) => m.y);

        this.markerContainer.addChild(this.polygonGfx);

        this.viewport.addChild(this.markerContainer);

        // Finish viewport setup
        this.app.stage.interactive = true;
        this.app.stage.addChild(this.viewport);

        this.viewport
            .drag()
            .pinch()
            .wheel();

        // Start rendering
        app.ticker.add(() => {
            this.config.stats?.begin();

            if (this.viewport.dirty) {
                this.draw();
                this.viewport.dirty = false;
            }

            this.config.stats?.end();
        });

        // Init handlers
        this.initHandlers();
    }

    private initHandlers() {
        // Viewport scale handlers
        let lastScale = this.viewport.scaled;
        let nextScale = lastScale;
        this.onScaleChange(lastScale, nextScale);

        let lastNativeZoom = this.nativeZoom;
        let nextNativeZoom = lastNativeZoom;
        this.onNativeZoomChange(lastNativeZoom, nextNativeZoom);

        this.viewport.on('zoomed-end', () => {
            // Scale change
            nextScale = this.viewport.scaled;
            if (lastScale != nextScale) {
                this.onScaleChange(lastScale, nextScale);
                lastScale = nextScale;
            }

            // Native zoom change
            nextNativeZoom = this.nativeZoom;
            if (lastNativeZoom != nextNativeZoom) {
                this.onNativeZoomChange(lastNativeZoom, nextNativeZoom);
                lastNativeZoom = nextNativeZoom;
            }
        });

        // Marker click handlers
        const $view = $(this.app.view);

        this.markerContainer.on('globalmousemove', (e) => {
            const local = this.markerContainer.toLocal(e.global);
            this.currentMarker = this.markerPoints.find(local.x, local.y, this.markerSize);
            if (!this.currentMarker) {
                $view.css('cursor', '');
                return
            }
            $view.css('cursor', 'pointer');


        });
        this.viewport.on('click', (e) => {
            if (this.currentMarker) {
                const local = this.markerContainer.toLocal(e.global);
                this.config.onMarkerClick(this.currentMarker, e.global, local);
            }
        });


        // Coordinate display handlers
        this.viewport.on('mousemove', (e) => {
            if (this.config.coordinateDisplay) {
                const markerLocal = this.markerContainer.toLocal(e.global);

                const tileLocal = this.tileContainer.toLocal(e.global);
                const currScale = Math.pow(2, this.config.maxNativeZoom - this.nativeZoom);
                const tileX = Math.floor(tileLocal.x / (this.config.tileSize.x / currScale));
                const tileY = Math.floor(tileLocal.y / (this.config.tileSize.y / currScale));

                this.config.coordinateDisplay.text(
                    `${markerLocal.x.toFixed(6)}, ${markerLocal.y.toFixed(6)} ｜ ${tileX}, ${tileY}`
                );
            }
        });
        this.viewport.on('mouseenter', () => this.config.coordinateDisplay?.show());
        this.viewport.on('mouseleave', () => this.config.coordinateDisplay?.hide());
    }

    public resize(width: number, height: number) {
        this.app.renderer.resize(width, height);
        this.viewport.resize(width, height);
    }

    private getTileTexture(tileCoord: Vec2): Texture | undefined {
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

    private getCurrentScale(): number {
        return 1 / Math.pow(2, this.config.maxNativeZoom - this.nativeZoom);
    }

    private getCurrentBounds(): Vec2 {
        return this.config.bounds.scale(this.getCurrentScale());
    }

    private getTilePos(tileCoord: Vec2): Vec2 {
        return tileCoord.mul(this.config.tileSize);
    }

    private getTileSprite(tileCoord: Vec2): Sprite | undefined {
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

    private eachTileInView(cb: (tileCoord: Vec2) => void) {
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

    public onScaleChange(_: number, newScale: number) {
        let l = Math.ceil((Math.log2(newScale)));
        this.nativeZoom = Math.min(Math.max(l, this.config.minNativeZoom), this.config.maxNativeZoom);
        this.drawMarkers();
    }
    public onNativeZoomChange(lastZoom: number, newZoom: number) {
        const ratio = Math.pow(2, newZoom) / Math.pow(2, lastZoom);
        this.viewport.center = new Point(this.viewport.center.x * ratio, this.viewport.center.y * ratio);
    }

    private drawMarkers() {
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

        this.markerSize = Math.max(0.15, (15 / this.config.crs.scale.x) / Math.pow(3, this.viewport.scaled));
        this.markerContainer.removeChildren(0);
        this.markerContainer.addChild(this.markerMesh.getMesh(this.markerSize));
        this.markerContainer.addChild(this.polygonGfx);
    }

    private draw() {
        // Clear viewport
        this.tileContainer.removeChildren(0);

        // Draw tiles
        this.eachTileInView((tileCoord: Vec2) => {
            const tileCoordWithZoom = tileCoord.withZ(tileCoord, this.nativeZoom);

            let sprite: Sprite | null = this.spriteCache.get(tileCoordWithZoom);
            if (!sprite) {
                // Create sprite
                sprite = this.getTileSprite(tileCoord) ?? null;
                if (sprite) {
                    this.spriteCache.set(tileCoordWithZoom, sprite);
                    this.tileContainer.addChild(sprite);
                }
            }
        });
    }

    public redraw(resetView = true) {
        this.viewport.worldWidth = this.config.tileSize.x * this.config.bounds.x;
        this.viewport.worldHeight = this.config.tileSize.y * this.config.bounds.y;
        this.draw();
        this.drawMarkers();

        if (resetView) {
            this.viewport.moveCorner(0, 0);
            this.viewport.setZoom(1);
        }
    }

    // TODO: marker should be generic or interface. WorldMap shouldn't know about D4 concerns.
    public addMarker(m: Marker) {
        this.markerMesh.addMarker(m.x, m.y, m.color);
        this.markerPoints.add(m);
    }

    public addPolygon(p: Array<Point>) {
        this.polygonGfx.lineStyle({
            width: 2,
            color: 0xaca491,
            alpha: 0.5,
        });
        this.polygonGfx.drawPolygon(p);
        this.polygonGfx.lineStyle();
    }

    public clearMarkers() {
        this.markerMesh.clear();
        this.markerContainer.removeChildren(0);
        this.markerPoints = quadtree<Marker>()
            .x(this.markerPoints.x())
            .y(this.markerPoints.y());
    }

    public clear() {
        this.clearMarkers();
        this.polygonGfx.clear();
        this.tileContainer.removeChildren(0);
    }
}