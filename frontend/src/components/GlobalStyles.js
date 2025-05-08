import { createGlobalStyle } from "styled-components";

export const GlobalStyles = createGlobalStyle`
  * {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }

  body {
    font-family: 'Poppins', sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    overflow-x: hidden;
  }

  button {
    font-family: 'Poppins', sans-serif;
    cursor: pointer;
    border: none;
    outline: none;
  }

  img {
    max-width: 100%;
    height: auto;
  }

  a {
    text-decoration: none;
    color: inherit;
  }

  /* VALORANT-inspired CSS variables */
  :root {
    --valorant-red: #FF4655;
    --valorant-blue: #0F1923;
    --valorant-white: #ECE8E1;
    --valorant-light-gray: #BDBCB7;
    --valorant-dark-gray: #292929;
    --valorant-green: #00FF84;
  }
`;
