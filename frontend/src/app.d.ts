declare module "*.svelte" {
  import type { SvelteComponentTyped } from "svelte";
  const component: typeof SvelteComponentTyped;
  export default component;
}

declare module "*.css" {
  const content: string;
  export default content;
}
