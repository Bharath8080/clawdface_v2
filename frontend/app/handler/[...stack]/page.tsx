import { StackHandler } from "@stackframe/stack";
import { stackServerApp } from "../../../stack";

export default async function Handler(props: { params: Promise<any>; searchParams: Promise<any> }) {
  const params = await props.params;
  const searchParams = await props.searchParams;

  return (
    <div className="min-h-screen w-full bg-[#050505] flex items-center justify-center p-4">
      <StackHandler 
        app={stackServerApp} 
        params={params} 
        searchParams={searchParams}
        fullPage
      />
    </div>
  );
}
