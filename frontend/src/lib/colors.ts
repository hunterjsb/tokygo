export const COLORS = {
  routes: {
    train: "#a78bfa",
    car: "#60a5fa",
    walk: "#34d399",
    flight: "#f87171",
    default: "#94a3b8",
  },
  locations: {
    hotel: "#60a5fa",
    airport: "#f87171",
    station: "#a78bfa",
    default: "#94a3b8",
  },
} as const;

export type RouteType = keyof typeof COLORS.routes;
export type LocationType = keyof typeof COLORS.locations;
