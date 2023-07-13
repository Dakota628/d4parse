document.addEventListener(
    'DOMContentLoaded',
    () => {
        loadGraph(graph => {
            // console.log(graph);
            // graph.forEachNode(node => {
            //     graph.setNodeAttribute(node, 'x', Math.random());
            //     graph.setNodeAttribute(node, 'y', Math.random());
            // });

            // graphologyLibrary.layout.circular.assign(graph);
            // graphologyLibrary.layoutForceAtlas2.assign(
            //     graph, {
            //         iterations: 300,
            //         settings: { ...graphologyLibrary.layoutForceAtlas2.inferSettings(graph), scalingRatio: 80 }
            //     }
            // );

            const container = document.getElementById("graph");
            const renderer = new Sigma(graph, container, {
                minCameraRatio: 0.0001,
                maxCameraRatio: 5,
            });
            const camera = renderer.getCamera();
        });
    },
    false,
);

function loadGraph(cb) {
    const graph = new graphology.Graph({
        type: 'directed',
    });
    const req = new XMLHttpRequest();
    req.open("GET", "../nodes.bin", true);
    req.responseType = "arraybuffer";
    req.onload = function (e) {
        const dv = new DataView(req.response);
        for (let p = 0; p < dv.byteLength; p += 12) {
            const id = dv.getInt32(p, true).toString();
            const x = dv.getFloat32(p + 4, true);
            const y = dv.getFloat32(p + 8, true);

            console.log(id, x, y);
            graph.addNode(id);
            graph.setNodeAttribute(id, 'x', x);
            graph.setNodeAttribute(id, 'y', y);
        }
        loadRefs(graph, cb);
    };
    req.send();
}

function loadRefs(graph, cb) {
    const req = new XMLHttpRequest();
    req.open("GET", "../refs.bin", true);
    req.responseType = "arraybuffer";
    req.onload = function (e) {
        const dv = new DataView(req.response);
        for (let p = 0; p < dv.byteLength; p += 8) {
            const from = dv.getInt32(p, true).toString();
            const to = dv.getInt32(p + 4, true).toString();
            graph.addEdge(from, to);
        }
        cb(graph);
    };
    req.send();
}