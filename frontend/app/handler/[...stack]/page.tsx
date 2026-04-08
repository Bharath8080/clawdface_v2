import { StackHandler } from "@stackframe/stack";
import { stackServerApp } from "../../../stack";

export default function Handler(props: any) {
  return (
    <div className="min-h-screen w-full bg-[#050505] flex items-center justify-center p-4">
      <StackHandler app={stackServerApp} {...props} />
    </div>
  );
}
