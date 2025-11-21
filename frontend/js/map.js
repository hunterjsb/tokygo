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
