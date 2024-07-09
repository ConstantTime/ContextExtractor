// src/components/CreateTicket.js
import axios from "axios";
import React, { useState } from "react";
import "./CreateTicket.css"; // We'll create this CSS file

function CreateTicket() {
  const [ticket, setTicket] = useState({
    customer_id: "",
    subject: "",
    content: "",
  });
  const [response, setResponse] = useState(null);

  const ticketData = {
    ...ticket,
    customer_id: parseInt(ticket.customer_id, 10),
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    axios
      .post("/api/tickets", ticketData)
      .then((response) => {
        setResponse(response.data);
        setTicket({ customer_id: "", subject: "", content: "" });
      })
      .catch((error) => console.error("Error creating ticket:", error));
  };

  const handleChange = (e) => {
    setTicket({ ...ticket, [e.target.name]: e.target.value });
  };

  return (
    <div className="create-ticket">
      <h2>Create Ticket</h2>
      <form onSubmit={handleSubmit}>
        <input
          type="number"
          name="customer_id"
          value={ticket.customer_id}
          onChange={handleChange}
          placeholder="Customer ID"
          required
        />
        <input
          type="text"
          name="subject"
          value={ticket.subject}
          onChange={handleChange}
          placeholder="Subject"
          required
        />
        <textarea
          name="content"
          value={ticket.content}
          onChange={handleChange}
          placeholder="Content"
          required
        />
        <button type="submit">Create Ticket</button>
      </form>
      {response && (
        <div className="response">
          <h3>Ticket Created</h3>
          <p className="status">
            Status: <span>{response.status}</span>
          </p>
          {response.email_draft && (
            <div className="email-draft">
              <h4>Email Draft (Context Demanded)</h4>
              <pre>{response.email_draft}</pre>
            </div>
          )}
          {response.context && (
            <p className="context">
              Context: <span>{response.context}</span>
            </p>
          )}
        </div>
      )}
    </div>
  );
}

export default CreateTicket;
