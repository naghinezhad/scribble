import htmx from "htmx.org";

export {}; // Ensure this file is treated as a module

declare global {
    interface Window {
        htmx: typeof htmx;
    }
}
