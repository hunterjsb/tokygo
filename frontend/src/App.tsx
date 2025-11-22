import { useState } from "react";
import { MapView } from "@/components/MapView";
import { InfoPanel } from "@/components/InfoPanel";
import { ThemeProvider } from "@/components/theme-provider";

function App() {
  const [routesVisible, setRoutesVisible] = useState(true);

  const toggleRoutes = () => {
    setRoutesVisible(!routesVisible);
  };

  return (
    <ThemeProvider defaultTheme="dark">
      <div className="relative h-screen w-screen overflow-hidden">
        <MapView routesVisible={routesVisible} />
        <InfoPanel routesVisible={routesVisible} onToggleRoutes={toggleRoutes} />
      </div>
    </ThemeProvider>
  );
}

export default App;
