import axios from "axios";
import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";

function TicketDetail() {
  const [ticket, setTicket] = useState(null);
  const { id } = useParams();

  useEffect(() => {
    axios
      .get(`/api/tickets/${id}`)
      .then((response) => setTicket(response.data))
      .catch((error) => console.error("Error fetching ticket:", error));
  }, [id]);

  if (!ticket) return <div>Loading...</div>;

  return (
    <div>
      <h1>{ticket.subject}</h1>
      <p>Content: {ticket.content}</p>
      <p>Extracted Context: {ticket.context}</p>
    </div>
  );
}

export default TicketDetail;
