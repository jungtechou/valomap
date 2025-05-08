import React from "react";
import styled from "styled-components";
import MapRoulette from "./components/MapRoulette";
import Header from "./components/Header";
import Footer from "./components/Footer";
import { GlobalStyles } from "./components/GlobalStyles";

const AppContainer = styled.div`
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: linear-gradient(135deg, #0f1923 0%, #1a2733 100%);
  color: #ece8e1;
`;

const MainContent = styled.main`
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1rem;
`;

function App() {
  return (
    <>
      <GlobalStyles />
      <AppContainer>
        <Header />
        <MainContent>
          <MapRoulette />
        </MainContent>
        <Footer />
      </AppContainer>
    </>
  );
}

export default App;
