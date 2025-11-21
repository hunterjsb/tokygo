// Global variables
let cityColors = {};
let cities = [];
let selectedLayer = null;
let geojsonLayer = null;
let hexagonsVisible = true;

// API base URL - gets replaced at build time from .env
const API_BASE_URL = "http://localhost:8080";

// Initialize the map centered on central Japan
const map = L.map("map").setView([35.0, 135.5], 8);

// Add a nice light-themed base map
L.tileLayer("https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png", {
  attribution: "© OpenStreetMap contributors © CARTO",
  maxZoom: 19,
}).addTo(map);

// Fetch cities data from API
fetch(`${API_BASE_URL}/api/cities`)
  .then((response) => response.json())
  .then((data) => {
    cities = data.cities;
    cityColors = data.colors;

    // Populate legend
    populateLegend();

    // Add city markers after data is loaded
    addCityMarkers();

    // Load GeoJSON after cities are loaded
    loadGeoJSON();
  })
  .catch((error) => {
    console.error("Error loading cities:", error);
    document.getElementById("loading").textContent =
      "Error loading cities: " + error.message;
  });

function populateLegend() {
  const legendItems = document.getElementById("legend-items");
  legendItems.innerHTML = "";

  cities.forEach((city) => {
    const color = cityColors[city.name];
    const item = document.createElement("div");
    item.className = "legend-item";
    item.innerHTML = `
      <div class="legend-color" style="background-color: ${color};"></div>
      <span>${city.name}</span>
    `;
    legendItems.appendChild(item);
  });
}

function addCityMarkers() {
  cities.forEach((city) => {
    const marker = L.circleMarker([city.lat, city.lng], {
      radius: 8,
      fillColor: cityColors[city.name],
      color: "#fff",
      weight: 2,
      opacity: 1,
      fillOpacity: 0.9,
    }).addTo(map);

    marker.bindPopup(
      `<strong>${city.name}</strong><br>${city.lat.toFixed(4)}, ${city.lng.toFixed(4)}`,
    );
  });
}

function toggleHexagons() {
  if (geojsonLayer) {
    if (hexagonsVisible) {
      map.removeLayer(geojsonLayer);
      hexagonsVisible = false;
      document.getElementById("toggle-btn").textContent = "Show Hexagons";
    } else {
      map.addLayer(geojsonLayer);
      hexagonsVisible = true;
      document.getElementById("toggle-btn").textContent = "Hide Hexagons";
    }
  }
}

function loadGeoJSON() {
  // Load and display the GeoJSON from API
  fetch(`${API_BASE_URL}/api/geojson`)
    .then((response) => {
      // Check cache status
      const cacheStatus = response.headers.get("X-Cache");
      if (cacheStatus) {
        console.log("GeoJSON cache status:", cacheStatus);
      }
      return response.json();
    })
    .then((data) => {
      document.getElementById("loading").style.display = "none";

      // Update info panel
      document.getElementById("cell-count").textContent = data.features.length;
      if (data.features.length > 0 && data.features[0].properties.resolution) {
        document.getElementById("resolution").textContent =
          data.features[0].properties.resolution;
      }

      // Add GeoJSON layer to map
      geojsonLayer = L.geoJSON(data, {
        style: function (feature) {
          const city = feature.properties.city || "Tokyo";
          const color = cityColors[city] || "#3388ff";

          return {
            fillColor: color,
            weight: 1,
            opacity: 0.6,
            color: "#666",
            fillOpacity: 0.25,
          };
        },
        onEachFeature: function (feature, layer) {
          const city = feature.properties.city || "Unknown";
          const baseColor = cityColors[city] || "#3388ff";

          // Add hover effect
          layer.on({
            mouseover: function (e) {
              const layer = e.target;
              if (layer !== selectedLayer) {
                layer.setStyle({
                  fillOpacity: 0.5,
                  weight: 2,
                });
              }
            },
            mouseout: function (e) {
              const layer = e.target;
              if (layer !== selectedLayer) {
                layer.setStyle({
                  fillOpacity: 0.25,
                  weight: 1,
                });
              }
            },
            click: function (e) {
              const layer = e.target;

              // Reset previously selected layer
              if (selectedLayer) {
                const prevCity =
                  selectedLayer.feature.properties.city || "Tokyo";
                const prevColor = cityColors[prevCity] || "#3388ff";
                selectedLayer.setStyle({
                  fillOpacity: 0.25,
                  weight: 1,
                  fillColor: prevColor,
                });
              }

              // Highlight selected layer
              layer.setStyle({
                fillOpacity: 0.7,
                weight: 3,
                fillColor: "#ff6b00",
              });

              selectedLayer = layer;

              // Update info panel
              const h3Index = feature.properties.h3_index || "Unknown";
              document.getElementById("selected-cell").textContent = h3Index;
              document.getElementById("selected-city").textContent =
                "📍 " + city;
              document.getElementById("selected-city").style.color = baseColor;
            },
          });

          // Add popup with H3 index and city
          if (feature.properties.h3_index) {
            layer.bindPopup(`
              <strong>${city}</strong><br>
              <strong>H3 Index:</strong><br>
              <span style="font-size: 11px;">${feature.properties.h3_index}</span>
            `);
          }
        },
      }).addTo(map);

      // Fit map to show all hexagons
      map.fitBounds(geojsonLayer.getBounds());
    })
    .catch((error) => {
      document.getElementById("loading").textContent =
        "Error loading GeoJSON: " + error.message;
      console.error("Error loading GeoJSON:", error);
    });
}
