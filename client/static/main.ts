import hover from "./hover"
// Need to import sse so it is available to reports.js
import sse from "../sse";

(window as any).sse = sse

hover();