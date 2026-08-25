import "./App.css";

function App() {
  return (
    <div className="app">
      <header className="hero">
        <h1>量潮众包管理云</h1>
        <p className="tagline">自营发销售众包管理工具</p>
        <p className="positioning">
          面向外部渠道与代理——量潮自营发单、统一验收、按标准结算，
          让销售执行从「靠人盯」变成「按标准管」。
        </p>
        <a
          className="cta"
          href="https://studio.crowd.cloud.quanttide.com"
          target="_blank"
          rel="noreferrer"
        >
          进入众包管理云 →
        </a>
      </header>

      <section className="features">
        <h2>功能介绍</h2>
        <p className="features-lead">
          量潮众包管理云是管理方的后台（不是交易平台）——管住任务、执行方与结算三件事。
        </p>
        <div className="cards">
          <div className="card">
            <h3>任务审核</h3>
            <p>验收准则清晰，才可发单——说不清验收的任务，不发。</p>
          </div>
          <div className="card">
            <h3>执行方管理</h3>
            <p>准入认证——接单前，先知道对方是谁。</p>
          </div>
          <div className="card">
            <h3>结算</h3>
            <p>验收通过，记录结算——钱从哪来、付给谁、付了多少，有据可查。</p>
          </div>
        </div>
      </section>

      <footer>
        <p>量潮众包管理云 · crowd.cloud.quanttide.com</p>
      </footer>
    </div>
  );
}

export default App;
