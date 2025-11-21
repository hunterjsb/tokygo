// Global state
let map;
let routesVisible = true;

// API base URL - gets replaced at build time from .env
const API_BASE_URL = "http://localhost:8080";

// Mapbox public token
mapboxgl.accessToken =
  "pk.eyJ1IjoiaHVudGVyanNiIiwiYSI6ImNtaThkeWo4dzBiOTkyd3Exb3FpdzdweWQifQ.e8TLoDm5tLdFSr0KTwEpLA";

// Initialize
document.addEventListener("DOMContentLoaded", () => {
  initMap();
  loadData();
});

function initMap() {
  map = new mapboxgl.Map({
    container: "map",
    style: "mapbox://styles/mapbox/dark-v11",
    center: [135.5, 35.0],
    zoom: 7,
    projection: "globe",
  });

  map.addControl(new mapboxgl.NavigationControl(), "top-right");
}

async function loadData() {
  try {
    // Load all data
    const [routesData, locationsData] = await Promise.all([
      fetch(`${API_BASE_URL}/api/routes/lines`).then((r) => r.json()),
      fetch(`${API_BASE_URL}/api/locations`).then((r) => r.json()),
    ]);

    map.on("load", () => {
      addRoutes(routesData);
      addLocations(locationsData);
      hideLoading();
    });
  } catch (error) {
    console.error("Error loading data:", error);
    showError(error.message);
  }
}

function addRoutes(data) {
  map.addSource("routes", {
    type: "geojson",
    data: data,
  });

  map.addLayer({
    id: "routes-line",
    type: "line",
    source: "routes",
    layout: {
      "line-join": "round",
      "line-cap": "round",
    },
    paint: {
      "line-color": [
        "match",
        ["get", "route_type"],
        "train",
        "#a78bfa",
        "car",
        "#60a5fa",
        "walk",
        "#34d399",
        "flight",
        "#f87171",
        "#94a3b8",
      ],
      "line-width": 3,
      "line-opacity": 0.9,
      "line-blur": 0.5,
    },
  });

  // Route click handler
  map.on("click", "routes-line", (e) => {
    const props = e.features[0].properties;
    const distanceKm = (props.distance / 1000).toFixed(1);
    const durationMin = Math.round(props.duration / 60);

    new mapboxgl.Popup()
      .setLngLat(e.lngLat)
      .setHTML(
        `
        <strong>${props.route_name}</strong><br>
        ${distanceKm} km · ${durationMin} min
      `,
      )
      .addTo(map);
  });

  map.on("mouseenter", "routes-line", () => {
    map.getCanvas().style.cursor = "pointer";
  });

  map.on("mouseleave", "routes-line", () => {
    map.getCanvas().style.cursor = "";
  });

  // Load and render entire hex grid
  loadHexGrid();
}

async function loadHexGrid() {
  const grids = {};
  let inflightController = null;
  // dynamic, formula-based resolution selection (no hardcoded map)

  // H3 resolution control with smooth mapping + light hysteresis
  let activeResolution = 3;
  let lastResUpdateZoom = 0;

  // Base resolution as a smooth function of zoom (tunable)
  // Rough guide: zoom 4->res3, z6->res4, z8->res6, z10->res7, z12->res8, z14->res9, z16->res10, z18->res11
  function computeBaseResolution(zoom) {
    // More aggressive increase at high zoom (smaller hexes sooner)
    // Example: z12->res8, z14->res9, z16->res10, z18->res11/12
    const res = Math.floor(0.9 * zoom - 2.2);
    return Math.max(3, Math.min(14, res));
  }

  // Apply hysteresis to avoid flicker when hovering around boundaries
  function updateActiveResolution(zoom) {
    const target = computeBaseResolution(zoom);
    if (target > activeResolution) {
      // require the zoom to move forward a bit before increasing res
      if (zoom - lastResUpdateZoom >= 0.35 || target - activeResolution >= 2) {
        activeResolution = target;
        lastResUpdateZoom = zoom;
      }
    } else if (target < activeResolution) {
      // require the zoom to move backward a bit before decreasing res
      if (lastResUpdateZoom - zoom >= 0.35 || activeResolution - target >= 2) {
        activeResolution = target;
        lastResUpdateZoom = zoom;
      }
    }
    return activeResolution;
  }

  // Removed preload loop; hex grid is fetched per-viewport on moveend using a computed resolution

  // Add source and layers
  map.addSource("hex-grid", {
    type: "geojson",
    promoteId: "h3_index",
    data: {
      type: "FeatureCollection",
      features: [],
    },
  });

  map.addLayer({
    id: "hex-grid",
    type: "fill",
    source: "hex-grid",
    paint: {
      "fill-color": "#fff",
      "fill-opacity": 0.01,
    },
  });

  map.addLayer({
    id: "hex-grid-outline",
    type: "line",
    source: "hex-grid",
    paint: {
      "line-color": "#fff",
      "line-width": 0.4,
      "line-opacity": 0.04,
    },
  });

  // Spotlight layers (overlay) - opacity driven by feature-state
  // Spotlight layers using filters (no feature-state)
  map.addLayer({
    id: "hex-spotlight-center",
    type: "fill",
    source: "hex-grid",
    paint: {
      "fill-color": "#ffffff",
      "fill-opacity": 0.35,
    },
    filter: ["in", ["get", "h3_index"], ["literal", []]],
  });

  map.addLayer({
    id: "hex-spotlight-ring",
    type: "fill",
    source: "hex-grid",
    paint: {
      "fill-color": "#ffffff",
      "fill-opacity": 0.18,
    },
    filter: ["in", ["get", "h3_index"], ["literal", []]],
  });

  // Switch grid based on zoom/move end
  map.on("moveend", async () => {
    const zoom = map.getZoom();
    const targetResolution = updateActiveResolution(zoom);

    // Fetch only the cells in the current viewport at the target resolution
    const bounds = map.getBounds();
    const url = `${API_BASE_URL}/api/h3/grid_window?minLat=${bounds.getSouth()}&minLng=${bounds.getWest()}&maxLat=${bounds.getNorth()}&maxLng=${bounds.getEast()}&resolution=${targetResolution}`;

    try {
      // cancel any in-flight request for smoother interaction
      if (inflightController) inflightController.abort();
      inflightController = new AbortController();

      const resp = await fetch(url, { signal: inflightController.signal });
      const data = await resp.json();

      const features = Object.entries(data.cells).map(
        ([h3Index, cellData]) => ({
          type: "Feature",
          id: h3Index,
          geometry: {
            type: "Polygon",
            coordinates: [cellData.boundary],
          },
          properties: {
            h3_index: h3Index,
            center: cellData.center,
            neighbors: cellData.neighbors || [],
          },
        }),
      );

      map.getSource("hex-grid").setData({
        type: "FeatureCollection",
        features,
      });
    } catch (e) {
      // noop on fetch errors to keep panning smooth
    }
  });

  // Spotlight using setFilter on center + ring (no feature-state)
  let rafHandle = null;

  function setSpotlight(centerId, ringIds) {
    const centerFilter = [
      "in",
      ["get", "h3_index"],
      ["literal", centerId ? [centerId] : []],
    ];
    const ringFilter = ["in", ["get", "h3_index"], ["literal", ringIds || []]];
    map.setFilter("hex-spotlight-center", centerFilter);
    map.setFilter("hex-spotlight-ring", ringFilter);
  }

  map.on("mousemove", (e) => {
    if (!map.getSource("hex-grid")) return;
    if (rafHandle) cancelAnimationFrame(rafHandle);
    rafHandle = requestAnimationFrame(() => {
      const feats = map.queryRenderedFeatures(e.point, {
        layers: ["hex-grid"],
      });
      if (!feats || feats.length === 0) {
        setSpotlight(null, []);
        return;
      }
      const f = feats[0];
      const centerId = f.properties && f.properties.h3_index;
      const neighbors = (f.properties && f.properties.neighbors) || [];
      setSpotlight(centerId, neighbors);
    });
  });

  map.on("mouseleave", () => {
    if (!map.getSource("hex-grid")) return;
    setSpotlight(null, []);
  });

  // Also trigger updates on zoom end for snappier resolution swaps
  map.on("zoomend", () => map.fire("moveend"));
  // Initial fetch at current zoom/viewport
  map.fire("moveend");
}

function addLocations(data) {
  map.addSource("locations", {
    type: "geojson",
    data: data,
  });

  // Location circles
  map.addLayer({
    id: "locations",
    type: "circle",
    source: "locations",
    paint: {
      "circle-radius": 6,
      "circle-color": [
        "match",
        ["get", "type"],
        "hotel",
        "#60a5fa",
        "airport",
        "#f87171",
        "station",
        "#a78bfa",
        "#94a3b8",
      ],
      "circle-stroke-width": 2,
      "circle-stroke-color": "#fff",
      "circle-opacity": 0.9,
    },
  });

  // Location click handler
  map.on("click", "locations", (e) => {
    const props = e.features[0].properties;
    const coords = e.features[0].geometry.coordinates.slice();

    new mapboxgl.Popup()
      .setLngLat(coords)
      .setHTML(
        `
        <strong>${props.name}</strong><br>
        <span style="text-transform: capitalize;">${props.type}</span> · ${props.city}
      `,
      )
      .addTo(map);
  });

  map.on("mouseenter", "locations", () => {
    map.getCanvas().style.cursor = "pointer";
  });

  map.on("mouseleave", "locations", () => {
    map.getCanvas().style.cursor = "";
  });
}

function toggleRoutes() {
  if (!map.getLayer("routes-line")) return;

  if (routesVisible) {
    map.setLayoutProperty("routes-line", "visibility", "none");
    document.getElementById("toggle-routes-btn").innerHTML =
      '<i class="fas fa-route"></i><span>Show Routes</span>';
    document.getElementById("toggle-routes-btn").classList.remove("active");
  } else {
    map.setLayoutProperty("routes-line", "visibility", "visible");
    document.getElementById("toggle-routes-btn").innerHTML =
      '<i class="fas fa-route"></i><span>Hide Routes</span>';
    document.getElementById("toggle-routes-btn").classList.add("active");
  }

  routesVisible = !routesVisible;
}

function hideLoading() {
  document.getElementById("loading").style.display = "none";
}

function showError(message) {
  const loading = document.getElementById("loading");
  loading.textContent = "Error: " + message;
  loading.style.color = "#f87171";
}
