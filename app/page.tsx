import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import Image from "next/image";

export default function Home() {
  return (
    <div>
      {/* indigo 靛蓝色 ; 500 色阶 */}
      <p className="test-3xl font-bold text-indigo-500">
        Hello Discord Clone
      </p>
      <Button>Click me</Button>
    </div>
  );
}
