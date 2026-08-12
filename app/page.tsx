import { Button } from "@base-ui/react";
import Image from "next/image";

export default function Home() {
  return (
    <div className="flex flex-col">
      <p className="text-2xl text-center">Welcome to My Next.js App!</p>
      <Button>Click me</Button>
    </div>
  );
}
