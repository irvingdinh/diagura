import "./styles/globals.css";

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <p>Lorem ipsum dolor sit amet</p>
  </StrictMode>,
);
