// src/components/TicketList.js
import axios from "axios";
import React, { useEffect, useState } from "react";

function TicketList() {
  const [tickets, setTickets] = useState([]);

  useEffect(() => {
    axios
      .get("/api/tickets")
      .then((response) => setTickets(response.data))
      .catch((error) => console.error("Error fetching tickets:", error));
  }, []);

  return (
    <div>
      <h2>Tickets</h2>
      {tickets.map((ticket) => (
        <div key={ticket.id} className="card">
          <h3>{ticket.subject}</h3>
          <p>
            <strong>Customer ID:</strong> {ticket.customer_id}
          </p>
          <p>
            <strong>Content:</strong> {ticket.content}
          </p>
          <p>
            <strong>Context:</strong> {ticket.context || "No context available"}
          </p>
          <p>
            <strong>Status:</strong> {ticket.status}
          </p>
        </div>
      ))}
    </div>
  );
}

export default TicketList;
