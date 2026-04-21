import { StackHandler } from "@stackframe/stack";

export default function Handler() {
  return (
    <div className="min-h-screen w-full bg-[#050505] flex items-center justify-center p-4">
      <StackHandler fullPage />
    </div>
  );
}
