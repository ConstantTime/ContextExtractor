import React from "react";
import {
  BrowserRouter as Router,
  NavLink,
  Route,
  Routes,
} from "react-router-dom";
import "./App.css";
import ContextRules from "./components/ContextRules";
import CreateTicket from "./components/CreateTicket";
import TicketList from "./components/TicketList";

function App() {
  return (
    <Router>
      <div className="App">
        <nav className="tabs">
          <NavLink
            to="/"
            end
            className={({ isActive }) => (isActive ? "active" : "")}
          >
            Tickets
          </NavLink>
          <NavLink
            to="/create-ticket"
            className={({ isActive }) => (isActive ? "active" : "")}
          >
            Create Ticket
          </NavLink>
          <NavLink
            to="/context-rules"
            className={({ isActive }) => (isActive ? "active" : "")}
          >
            Context Rules
          </NavLink>
        </nav>

        <Routes>
          <Route path="/" element={<TicketList />} />
          <Route path="/create-ticket" element={<CreateTicket />} />
          <Route path="/context-rules" element={<ContextRules />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
