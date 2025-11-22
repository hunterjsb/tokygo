import {
  Route,
  Hotel,
  Plane,
  TrainFront,
  Car,
  Moon,
  Sun,
  Menu,
  X,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import { useTheme } from "@/components/theme-provider";
import { COLORS } from "@/lib/colors";
import { useState } from "react";

interface InfoPanelProps {
  routesVisible: boolean;
  onToggleRoutes: () => void;
}

export function InfoPanel({ routesVisible, onToggleRoutes }: InfoPanelProps) {
  const { theme, setTheme } = useTheme();
  const [isOpen, setIsOpen] = useState(false);

  const toggleTheme = () => {
    setTheme(theme === "dark" ? "light" : "dark");
  };

  return (
    <>
      {/* Mobile hamburger button */}
      <Button
        variant="default"
        size="icon"
        onClick={() => setIsOpen(!isOpen)}
        className="md:hidden absolute left-4 top-4 z-[1001] shadow-lg"
      >
        {isOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
      </Button>

      {/* Panel */}
      <Card
        className={`
        absolute left-4 top-4 z-[1000] shadow-lg backdrop-blur-sm bg-card/95 pointer-events-auto
        w-[calc(100vw-2rem)] max-w-80
        transition-transform duration-300 ease-in-out
        md:translate-x-0
        ${isOpen ? "translate-x-0" : "-translate-x-[calc(100%+2rem)] md:translate-x-0"}
      `}
      >
        <CardHeader className="pb-4">
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <TrainFront className="h-6 w-6" />
              <span className="hidden sm:inline">Japan Travel</span>
              <span className="sm:hidden">Tokygo</span>
            </CardTitle>
            <Button
              variant="ghost"
              size="icon"
              onClick={toggleTheme}
              className="h-8 w-8"
            >
              {theme === "dark" ? (
                <Sun className="h-4 w-4" />
              ) : (
                <Moon className="h-4 w-4" />
              )}
            </Button>
          </div>
        </CardHeader>

        <CardContent className="space-y-4">
          {/* Routes Toggle */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Route className="h-4 w-4" />
              <span className="text-sm font-medium">Show Routes</span>
            </div>
            <Switch checked={routesVisible} onCheckedChange={onToggleRoutes} />
          </div>

          <Separator />

          {/* Routes Legend */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Route className="h-4 w-4" />
              <span className="text-sm font-semibold">Routes</span>
            </div>
            <div className="space-y-2 pl-6">
              <LegendItem
                color={COLORS.routes.train}
                label="Shinkansen"
                icon={<TrainFront className="h-3 w-3" />}
              />
              <LegendItem
                color={COLORS.routes.car}
                label="Car"
                icon={<Car className="h-3 w-3" />}
              />
            </div>
          </div>

          <Separator />

          {/* Locations Legend */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Badge variant="outline" className="gap-1">
                <span className="text-xs font-semibold">Locations</span>
              </Badge>
            </div>
            <div className="space-y-2 pl-6">
              <LegendItem
                color={COLORS.locations.hotel}
                label="Hotels"
                icon={<Hotel className="h-3 w-3" />}
                isCircle
              />
              <LegendItem
                color={COLORS.locations.airport}
                label="Airports"
                icon={<Plane className="h-3 w-3" />}
                isCircle
              />
              <LegendItem
                color={COLORS.locations.station}
                label="Stations"
                icon={<TrainFront className="h-3 w-3" />}
                isCircle
              />
            </div>
          </div>
        </CardContent>
      </Card>
    </>
  );
}

interface LegendItemProps {
  color: string;
  label: string;
  icon?: React.ReactNode;
  isCircle?: boolean;
}

function LegendItem({ color, label, icon, isCircle = false }: LegendItemProps) {
  return (
    <div className="flex items-center gap-2">
      <div
        className={`${isCircle ? "h-3 w-3 rounded-full border-2 border-white" : "h-1 w-6 rounded-full"}`}
        style={{ backgroundColor: color }}
      />
      <span className="text-sm text-muted-foreground">{label}</span>
      {icon && <span className="text-muted-foreground">{icon}</span>}
    </div>
  );
}
