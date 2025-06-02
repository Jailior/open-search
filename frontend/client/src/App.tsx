  import SearchPage from "./SearchPage";

  function App() {
    return (
      <div className="min-h-screen flex flex-col bg-white text-gray-900">

        <main className="flex-grow flex flex-col justify-center">
          <SearchPage />
        </main>

        <footer className="bg-white text-center text-sm text-black-eerie py-4">
          Powered by{" "}
          <span className="font-semibold text-green-reseda">Ali Osman</span> • {" "}
          <a
            href="https://github.com/Jailior/open-search"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:underline text-green-reseda hover:text-green-celadon-50"
          >
            GitHub
          </a>{" "}
          |{" "}
          <a
            href="https://linkedin.com/in/ali-osman1"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:underline text-green-reseda hover:text-green-celadon-50"
          >
            LinkedIn
          </a>
        </footer>
      </div>
    );
  }

  export default App;