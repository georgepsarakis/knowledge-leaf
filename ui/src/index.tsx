import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import EventsOnThisDay from "./OnThisDay";
import {BrowserRouter, Route, Routes} from "react-router";

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);
root.render(
  <React.StrictMode>
      <BrowserRouter>
          <Routes>
              <Route path="/" element={<App />} />
              <Route path="/on-this-day/events" element={<EventsOnThisDay/>} />
          </Routes>
      </BrowserRouter>
  </React.StrictMode>
);
