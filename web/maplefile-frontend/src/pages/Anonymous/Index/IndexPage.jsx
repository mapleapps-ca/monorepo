// monorepo/web/maplefile-frontend/src/pages/Anonymous/Index/IndexPage.jsx
import { useState } from "react";
import { Link } from "react-router";
// import "./App.css";

function IndexPage() {
  return (
    <>
      <h1>Index Page</h1>
      <div>
        <p>Welcome to MapleApps</p>
        <div>
          <Link to="/register">
            <button>Register</button>
          </Link>
        </div>
      </div>
    </>
  );
}

export default IndexPage;
