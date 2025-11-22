import { useEffect, useRef, useState } from "react";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import { COLORS } from "@/lib/colors";
import { useTheme } from "@/components/theme-provider";

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

interface MapViewProps {
  routesVisible: boolean;
}

export function MapView({ routesVisible }: MapViewProps) {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<L.Map | null>(null);
  const tileLayer = useRef<L.TileLayer | null>(null);
  const routesLayer = useRef<L.LayerGroup | null>(null);
  const locationsLayer = useRef<L.LayerGroup | null>(null);
  const hexGridLayer = useRef<L.LayerGroup | null>(null);
  const [error, setError] = useState<string | null>(null);
  const { theme } = useTheme();

  useEffect(() => {
    if (!mapContainer.current || map.current) return;

    // Initialize Leaflet map
    map.current = L.map(mapContainer.current, {
      center: [35.0, 135.5],
      zoom: 7,
      zoomControl: false,
      touchZoom: true,
      scrollWheelZoom: true,
      doubleClickZoom: true,
    });

    // Add CARTO tiles (free, no token required)
    const isDark =
      theme === "dark" ||
      (theme === "system" &&
        window.matchMedia("(prefers-color-scheme: dark)").matches);
    const tileUrl = isDark
      ? "https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}.png"
      : "https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png";

    tileLayer.current = L.tileLayer(tileUrl, {
      attribution:
        '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
      maxZoom: 19,
    }).addTo(map.current);

    // Add zoom control - bottom-right on mobile, top-right on desktop
    const zoomPosition = window.innerWidth < 768 ? "bottomright" : "topright";
    L.control.zoom({ position: zoomPosition }).addTo(map.current);

    // Initialize layer groups
    routesLayer.current = L.layerGroup().addTo(map.current);
    locationsLayer.current = L.layerGroup().addTo(map.current);
    hexGridLayer.current = L.layerGroup().addTo(map.current);

    // Load data
    async function loadData() {
      try {
        const [routesData, locationsData] = await Promise.all([
          fetch(`${API_BASE_URL}/api/routes/lines`).then((r) => r.json()),
          fetch(`${API_BASE_URL}/api/locations`).then((r) => r.json()),
        ]);

        addRoutes(routesData);
        addLocations(locationsData);
        loadHexGrid();
      } catch (err) {
        console.error("Error loading data:", err);
        setError(err instanceof Error ? err.message : "Failed to load data");
      }
    }

    loadData();

    // Cleanup
    return () => {
      map.current?.remove();
      map.current = null;
    };
  }, [theme]);

  // Update map tiles when theme changes
  useEffect(() => {
    if (!map.current || !tileLayer.current) return;

    const isDark =
      theme === "dark" ||
      (theme === "system" &&
        window.matchMedia("(prefers-color-scheme: dark)").matches);
    const tileUrl = isDark
      ? "https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}.png"
      : "https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png";

    tileLayer.current.setUrl(tileUrl);
  }, [theme]);

  // Handle route visibility toggle
  useEffect(() => {
    if (!map.current || !routesLayer.current) return;

    if (routesVisible) {
      routesLayer.current.addTo(map.current);
    } else {
      routesLayer.current.remove();
    }
  }, [routesVisible]);

  function addRoutes(data: any) {
    if (!map.current || !routesLayer.current) return;

    data.features.forEach((feature: any) => {
      const coords = feature.geometry.coordinates.map((coord: number[]) => [
        coord[1],
        coord[0],
      ]);
      const props = feature.properties;

      const routeType = props.route_type || "default";
      const color =
        COLORS.routes[routeType as keyof typeof COLORS.routes] ||
        COLORS.routes.default;

      const polyline = L.polyline(coords, {
        color: color,
        weight: 3,
        opacity: 0.9,
      });

      const distanceKm = (props.distance / 1000).toFixed(1);
      const durationMin = Math.round(props.duration / 60);

      polyline.bindPopup(`
        <strong>${props.route_name}</strong><br>
        ${distanceKm} km · ${durationMin} min
      `);

      polyline.addTo(routesLayer.current!);
    });
  }

  function addLocations(data: any) {
    if (!map.current || !locationsLayer.current) return;

    data.features.forEach((feature: any) => {
      const coords: [number, number] = [
        feature.geometry.coordinates[1],
        feature.geometry.coordinates[0],
      ];
      const props = feature.properties;

      const locationType = props.type || "default";
      const color =
        COLORS.locations[locationType as keyof typeof COLORS.locations] ||
        COLORS.locations.default;

      const circleMarker = L.circleMarker(coords, {
        radius: 6,
        fillColor: color,
        color: "#fff",
        weight: 2,
        opacity: 0.9,
        fillOpacity: 0.9,
      });

      circleMarker.bindPopup(`
        <strong>${props.name}</strong><br>
        <span style="text-transform: capitalize;">${props.type}</span> · ${props.city}
      `);

      circleMarker.addTo(locationsLayer.current!);
    });
  }

  function loadHexGrid() {
    if (!map.current || !hexGridLayer.current) return;

    let activeResolution = 3;
    let lastResUpdateZoom = 0;
    let inflightController: AbortController | null = null;
    let updateTimeout: NodeJS.Timeout | null = null;

    function computeBaseResolution(zoom: number): number {
      const res = Math.floor(0.9 * zoom - 2.2);
      return Math.max(3, Math.min(14, res));
    }

    function updateActiveResolution(zoom: number): number {
      const target = computeBaseResolution(zoom);
      if (target > activeResolution) {
        if (
          zoom - lastResUpdateZoom >= 0.35 ||
          target - activeResolution >= 2
        ) {
          activeResolution = target;
          lastResUpdateZoom = zoom;
        }
      } else if (target < activeResolution) {
        if (
          lastResUpdateZoom - zoom >= 0.35 ||
          activeResolution - target >= 2
        ) {
          activeResolution = target;
          lastResUpdateZoom = zoom;
        }
      }
      return activeResolution;
    }

    async function updateHexGrid() {
      if (!map.current || !hexGridLayer.current) return;

      const zoom = map.current.getZoom();
      const targetResolution = updateActiveResolution(zoom);

      const bounds = map.current.getBounds();
      const url = `${API_BASE_URL}/api/h3/grid_window?minLat=${bounds.getSouth()}&minLng=${bounds.getWest()}&maxLat=${bounds.getNorth()}&maxLng=${bounds.getEast()}&resolution=${targetResolution}`;

      try {
        if (inflightController) inflightController.abort();
        inflightController = new AbortController();

        const resp = await fetch(url, { signal: inflightController.signal });
        const data = await resp.json();

        // Clear existing hex grid
        hexGridLayer.current!.clearLayers();

        // Add hex cells
        Object.entries(data.cells).forEach(
          ([h3Index, cellData]: [string, any]) => {
            const coords = cellData.boundary.map((coord: number[]) => [
              coord[1],
              coord[0],
            ]);

            const polygon = L.polygon(coords, {
              color: "#fff",
              weight: 0.4,
              opacity: 0.04,
              fillColor: "#fff",
              fillOpacity: 0.01,
            });

            polygon.on("mouseover", () => {
              polygon.setStyle({
                fillOpacity: 0.35,
              });

              // Highlight neighbors
              const neighbors = cellData.neighbors || [];
              hexGridLayer.current?.eachLayer((layer: any) => {
                const layerH3 = layer.options.h3Index;
                if (neighbors.includes(layerH3)) {
                  layer.setStyle({
                    fillOpacity: 0.18,
                  });
                }
              });
            });

            polygon.on("mouseout", () => {
              polygon.setStyle({
                fillOpacity: 0.01,
              });

              // Reset neighbors
              hexGridLayer.current?.eachLayer((layer: any) => {
                layer.setStyle({
                  fillOpacity: 0.01,
                });
              });
            });

            (polygon as any).options.h3Index = h3Index;
            polygon.addTo(hexGridLayer.current!);
          },
        );
      } catch (e) {
        // Ignore fetch errors for smooth panning
      }
    }

    // Debounced update handler
    const debouncedUpdate = () => {
      if (updateTimeout) clearTimeout(updateTimeout);
      updateTimeout = setTimeout(() => {
        updateHexGrid();
      }, 150);
    };

    map.current.on("moveend", debouncedUpdate);
    map.current.on("zoomend", debouncedUpdate);
    updateHexGrid();
  }

  if (error) {
    return (
      <div className="flex h-full w-full items-center justify-center text-destructive">
        Error: {error}
      </div>
    );
  }

  return <div ref={mapContainer} className="h-full w-full" />;
}
