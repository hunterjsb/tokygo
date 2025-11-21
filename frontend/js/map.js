// Global variables
let cityColors = {};
let cities = [];
let routesVisible = true;
let map;

// API base URL - gets replaced at build time from .env
const API_BASE_URL = "http://localhost:8080";

// Mapbox public token (safe for frontend, used for map display)
mapboxgl.accessToken =
  "pk.eyJ1IjoiaHVudGVyanNiIiwiYSI6ImNtaThkeWo4dzBiOTkyd3Exb3FpdzdweWQifQ.e8TLoDm5tLdFSr0KTwEpLA";

// Fetch cities data from API
fetch(`${API_BASE_URL}/api/cities`)
  .then((response) => response.json())
  .then((data) => {
    cities = data.cities;
    cityColors = data.colors;

    // Populate legend
    populateLegend();

    // Initialize map after cities are loaded
    initializeMap();

    // Load routes by default
    loadRoutes();

    // Hide loading indicator
    document.getElementById("loading").style.display = "none";
  })
  .catch((error) => {
    console.error("Error loading cities:", error);
    document.getElementById("loading").textContent =
      "Error loading cities: " + error.message;
  });

function populateLegend() {
  const legendItems = document.getElementById("legend-items");
  legendItems.innerHTML = "";

  const cityIcons = {
    Tokyo: "fa-building",
    Kyoto: "fa-torii-gate",
    Osaka: "fa-landmark",
  };

  cities.forEach((city) => {
    const item = document.createElement("div");
    item.className = "legend-item";
    item.innerHTML = `
      <i class="fas ${cityIcons[city.name] || "fa-map-marker-alt"} legend-icon"></i>
      <span>${city.name}</span>
    `;
    legendItems.appendChild(item);
  });
}

function togglePanel() {
  const panel = document.getElementById("info-panel");
  const content = document.getElementById("panel-content");
  const icon = document.getElementById("collapse-icon");

  panel.classList.toggle("collapsed");

  if (panel.classList.contains("collapsed")) {
    icon.classList.remove("fa-chevron-down");
    icon.classList.add("fa-chevron-up");
  } else {
    icon.classList.remove("fa-chevron-up");
    icon.classList.add("fa-chevron-down");
  }
}

function initializeMap() {
  // Initialize Mapbox map with sleek dark style
  map = new mapboxgl.Map({
    container: "map",
    style: "mapbox://styles/mapbox/dark-v11",
    center: [135.5, 35.0],
    zoom: 7,
    projection: "globe",
  });

  // Add navigation controls
  map.addControl(new mapboxgl.NavigationControl(), "top-right");

  // Add city markers with Font Awesome icons
  const cityIcons = {
    Tokyo: "fa-building",
    Kyoto: "fa-torii-gate",
    Osaka: "fa-landmark",
  };

  cities.forEach((city) => {
    const el = document.createElement("div");
    el.className = "city-marker";
    el.innerHTML = `<i class="fas ${cityIcons[city.name] || "fa-map-marker-alt"}"></i>`;

    new mapboxgl.Marker({ element: el, anchor: "bottom", draggable: false })
      .setLngLat([city.lng, city.lat])
      .setPopup(
        new mapboxgl.Popup({ offset: 25 }).setHTML(
          `<div style="text-align: center;">
            <i class="fas ${cityIcons[city.name]} fa-2x" style="margin-bottom: 8px; color: #fff;"></i><br>
            <strong>${city.name}</strong>
          </div>`,
        ),
      )
      .addTo(map);
  });
}

function toggleRoutes() {
  if (routesVisible) {
    if (map.getLayer("routes-line")) {
      map.setLayoutProperty("routes-line", "visibility", "none");
    }
    routesVisible = false;
    document.getElementById("toggle-routes-btn").innerHTML =
      '<i class="fas fa-route"></i><span>Show Routes</span>';
    document.getElementById("toggle-routes-btn").classList.remove("active");
  } else {
    if (map.getLayer("routes-line")) {
      map.setLayoutProperty("routes-line", "visibility", "visible");
    } else {
      loadRoutes();
    }
    routesVisible = true;
    document.getElementById("toggle-routes-btn").innerHTML =
      '<i class="fas fa-route"></i><span>Hide Routes</span>';
    document.getElementById("toggle-routes-btn").classList.add("active");
  }
}

function loadRoutes() {
  fetch(`${API_BASE_URL}/api/routes/lines`)
    .then((response) => response.json())
    .then((data) => {
      console.log(`Loaded ${data.features.length} route lines`);

      // Add route lines as a source
      map.addSource("routes", {
        type: "geojson",
        data: data,
      });

      // Add line layer
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
            "#a78bfa", // purple - shinkansen
            "car",
            "#60a5fa", // blue - car
            "walk",
            "#34d399", // green - walk
            "flight",
            "#f87171", // red - flight
            "#94a3b8", // gray - default
          ],
          "line-width": 3,
          "line-opacity": 0.9,
          "line-blur": 0.5,
        },
      });

      // Add click handler for popups
      map.on("click", "routes-line", (e) => {
        const feature = e.features[0];
        const routeName = feature.properties.route_name || "Unknown Route";
        const routeType = feature.properties.route_type || "unknown";
        const distanceKm = (feature.properties.distance / 1000).toFixed(1);
        const durationMin = Math.round(feature.properties.duration / 60);

        const typeIcons = {
          train: "fa-train",
          car: "fa-car",
          walk: "fa-walking",
          flight: "fa-plane",
        };

        const typeLabels = {
          train: "Shinkansen",
          car: "Car",
          walk: "Walking",
          flight: "Flight",
        };

        new mapboxgl.Popup()
          .setLngLat(e.lngLat)
          .setHTML(
            `
            <strong>${routeName}</strong><br>
            <div style="display: flex; align-items: center; gap: 6px; margin: 8px 0;">
              <i class="fas ${typeIcons[routeType] || "fa-map-marker-alt"}" style="width: 16px;"></i>
              <span>${typeLabels[routeType] || routeType}</span>
            </div>
            Distance: ${distanceKm} km<br>
            Duration: ${durationMin} min
          `,
          )
          .addTo(map);
      });

      // Change cursor on hover
      map.on("mouseenter", "routes-line", () => {
        map.getCanvas().style.cursor = "pointer";
      });

      map.on("mouseleave", "routes-line", () => {
        map.getCanvas().style.cursor = "";
      });

      routesVisible = true;
      document.getElementById("toggle-routes-btn").innerHTML =
        '<i class="fas fa-route"></i><span>Hide Routes</span>';
      document.getElementById("toggle-routes-btn").classList.add("active");
    })
    .catch((error) => {
      console.error("Error loading routes:", error);
      document.getElementById("toggle-routes-btn").textContent =
        "Error Loading Routes";
    });
}
