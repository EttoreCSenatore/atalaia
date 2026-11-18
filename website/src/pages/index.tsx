import type {ReactNode} from 'react';
import Link from '@docusaurus/Link';
import Translate, {translate} from '@docusaurus/Translate';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';

import styles from './index.module.css';

type Feature = {title: string; body: string};

const features: Feature[] = [
  {
    title: translate({id: 'home.feature.0.title', message: 'Descoberta de dispositivos'}),
    body: translate({id: 'home.feature.0.body', message: 'Encontra celulares, computadores, TVs e aparelhos da sua rede (ARP, mDNS, SSDP) e identifica fabricante, nome e tipo.'}),
  },
  {
    title: translate({id: 'home.feature.1.title', message: 'Histórico e alertas'}),
    body: translate({id: 'home.feature.1.body', message: 'Linha do tempo online/offline e aviso na hora em que um dispositivo desconhecido entra, por ntfy, Telegram, e-mail ou webhook.'}),
  },
  {
    title: translate({id: 'home.feature.2.title', message: 'Verificações de segurança defensivas'}),
    body: translate({id: 'home.feature.2.body', message: 'Detecta ARP spoofing e DHCP falso, serviços abertos, painéis de administração expostos, TLS fraco e CVEs conhecidas, com dica de correção.'}),
  },
  {
    title: translate({id: 'home.feature.3.title', message: 'Ferramentas de rede'}),
    body: translate({id: 'home.feature.3.body', message: 'Ping, traceroute, DNS, Wake-on-LAN, teste de velocidade e monitoramento da qualidade da internet no mesmo lugar.'}),
  },
  {
    title: translate({id: 'home.feature.4.title', message: 'Seguro por padrão'}),
    body: translate({id: 'home.feature.4.body', message: 'Login obrigatório, só na rede local, varreduras limitadas às suas redes, verificações ativas opcionais e auditadas, dados na sua máquina.'}),
  },
  {
    title: translate({id: 'home.feature.5.title', message: 'Software livre'}),
    body: translate({id: 'home.feature.5.body', message: 'Um binário em Go ou imagem Docker com o app web dentro. AGPL-3.0-or-later, documentação bilíngue, sem conta em nuvem.'}),
  },
];

const roles = [
  translate({id: 'home.role.0', message: 'Laboratórios caseiros'}),
  translate({id: 'home.role.1', message: 'Pequenos escritórios'}),
  translate({id: 'home.role.2', message: 'Segurança e DevOps'}),
  translate({id: 'home.role.3', message: 'Curiosos'}),
];

export default function Home(): ReactNode {
  return (
    <Layout
      title={translate({id: 'home.title', message: 'Vigie a sua rede'})}
      description={translate({
        id: 'home.description',
        message: 'Atalaia é um vigia de rede livre e de código aberto: descubra dispositivos, acompanhe o histórico, receba alertas e faça verificações de segurança defensivas. App web, auto-hospedado.',
      })}>
      <header className={styles.hero}>
        <div className="container">
          <Heading as="h1" className={styles.title}>
            Atalaia
          </Heading>
          <p className={styles.tagline}>
            <Translate id="home.tagline">Vigie a sua rede: descubra dispositivos, acompanhe o histórico, receba alertas e faça verificações de segurança defensivas. Livre, aberto e auto-hospedado.</Translate>
          </p>
          <div className={styles.roles}>
            {roles.map((r) => (
              <span key={r} className={styles.role}>
                {r}
              </span>
            ))}
          </div>
          <div className={styles.buttons}>
            <Link className="button button--primary button--lg" to="/docs/intro">
              <Translate id="home.docs">Ler a documentação</Translate>
            </Link>
            <Link className="button button--secondary button--lg" to="/docs/roadmap">
              <Translate id="home.roadmap">Roadmap</Translate>
            </Link>
          </div>
        </div>
      </header>
      <main className="container">
        <section className={styles.features}>
          {features.map((f) => (
            <div key={f.title} className={styles.card}>
              <Heading as="h3">{f.title}</Heading>
              <p>{f.body}</p>
            </div>
          ))}
        </section>
      </main>
    </Layout>
  );
}
