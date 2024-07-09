import axios from "axios";
import React, { useEffect, useState } from "react";

function ContextRules() {
  const [rules, setRules] = useState([]);
  const [newRule, setNewRule] = useState({ name: "", keywords: "" });

  useEffect(() => {
    axios
      .get("/api/context-rules")
      .then((response) => setRules(response.data.rules))
      .catch((error) => console.error("Error fetching context rules:", error));
  }, []);

  const handleSubmit = (e) => {
    e.preventDefault();
    axios
      .post("/api/context-rules", newRule)
      .then((response) => {
        setRules([...rules, response.data]);
        setNewRule({ name: "", keywords: "" });
      })
      .catch((error) => console.error("Error creating context rule:", error));
  };

  const handleChange = (e) => {
    setNewRule({ ...newRule, [e.target.name]: e.target.value });
  };

  return (
    <div>
      <h2>Context Rules</h2>
      {rules.map((rule) => (
        <div key={rule.id} className="card">
          <h3>{rule.name}</h3>
          <p>
            <strong>Keywords:</strong> {rule.keywords}
          </p>
        </div>
      ))}
      <div className="card">
        <h3>Add New Rule</h3>
        <form onSubmit={handleSubmit}>
          <input
            type="text"
            name="name"
            value={newRule.name}
            onChange={handleChange}
            placeholder="Rule Name"
            required
          />
          <input
            type="text"
            name="keywords"
            value={newRule.keywords}
            onChange={handleChange}
            placeholder="Keywords (comma-separated)"
            required
          />
          <button type="submit">Add Rule</button>
        </form>
      </div>
    </div>
  );
}

export default ContextRules;
