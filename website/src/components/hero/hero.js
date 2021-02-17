import React from "react";
import useBaseUrl from "@docusaurus/useBaseUrl";
import Link from "@docusaurus/Link";
import "./hero.css";
import FadeIn from "react-fade-in";

export default function hero() {
  return (
    <div className="main">
      <FadeIn transitionDuration="500">
        <div className="main__container">
          <div className="main__content">
            <h1>GraphQL Link</h1>
            <h2></h2>
            <p>is a GraphQL Gateway that lets proxy to other graphQL servers</p>
            <button className="main__btn">
              <Link to={useBaseUrl("docs/")}>Get Started</Link>
            </button>
          </div>

          <div className="main__img--container">
            <img
              id="main__img"
              src="./img/logo_without_text_three.png"
              alt="logo"
            />
          </div>
        </div>
      </FadeIn>
    </div>
  );
}
