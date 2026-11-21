import { useCallback, useEffect, useState } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter, Link, Route, Routes, useLocation, useParams } from 'react-router';
import { AppHeader, Button, Icon, PageHeading, Shell, buttonClass } from '@home-apps/ui';
import { request, type RationState, type Reload } from './api';
import { Day, Plans } from './day';
import { Editor } from './editor';
import { EditProduct, Products } from './products';
import './styles.css';

interface ScreenProps { data: RationState; reload: Reload; stale: boolean }

function MissingPage() {
  return <div className="empty">
    <PageHeading>Страница не найдена</PageHeading>
    <Link to="/" className={buttonClass('primary')}>К дневному рациону</Link>
  </div>;
}

function PlanScreen({ data, reload, stale, mode }: ScreenProps & { mode: 'view' | 'edit' | 'copy' }) {
  const { id } = useParams();
  const plan = data.plans.find((plan) => String(plan.id) === id);
  if (!plan) return stale ? null : <MissingPage />;
  return mode === 'view'
    ? <Day key={plan.id} plan={plan} plans={data.plans} reload={reload} />
    : <Editor key={mode + ':' + plan.id} mode={mode} plan={plan} products={data.products} reload={reload} />;
}

function ProductScreen({ data, reload, stale }: ScreenProps) {
  const { id } = useParams();
  const product = data.products.find((product) => String(product.id) === id);
  if (!product) return stale ? null : <MissingPage />;
  return <EditProduct key={product.id} product={product} reload={reload} />;
}

function App() {
  const { pathname } = useLocation();
  const [data, setData] = useState<RationState>();
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const stale = loading || !!error;
  const activePlan = data?.plans.find((plan) => plan.active);
  const selectedID = pathname.match(/^\/plans\/(\d+)$/)?.[1];
  const selected = pathname === '/' ? activePlan : data?.plans.find((plan) => String(plan.id) === selectedID);
  const section = pathname === '/' || selectedID ? 'day' : pathname.startsWith('/plans') ? 'plans' : pathname.startsWith('/products') ? 'products' : '';

  const reload = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      setData(await request<RationState>('/state'));
      return true;
    } catch (error) {
      setError((error as Error).message);
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void reload(); }, [reload]);

  return <Shell theme="ration">
    <AppHeader
      brand={<Link to="/"><Icon name="meal" />Рацион</Link>}
      navigation={<>
        <Link to="/" aria-current={section === 'day' ? 'page' : undefined}>День</Link>
        <Link to="/plans" aria-current={section === 'plans' ? 'page' : undefined}>Версии</Link>
        <Link to="/products" aria-current={section === 'products' ? 'page' : undefined}>Продукты</Link>
      </>}
      actions={selected && <Link className={buttonClass('quiet')} to={'/plans/' + selected.id + '/edit'}><Icon name="pencil" />Изменить</Link>}
    />
    <main className="ui-main ui-container" id="content" tabIndex={-1}>
      {error && <div role="alert" className="page-error">
        <p className="ui-error">{error}</p>
        <Button onClick={() => void reload()} disabled={loading}>Повторить</Button>
      </div>}
      {!data && loading && <p className="ui-muted" role="status">Загружаю рацион…</p>}
      {data && <Routes>
        <Route path="/" element={<Day key={activePlan?.id} plan={activePlan} plans={data.plans} reload={reload} />} />
        <Route path="/plans" element={<Plans plans={data.plans} reload={reload} />} />
        <Route path="/plans/new" element={<Editor mode="new" products={data.products} reload={reload} />} />
        <Route path="/plans/:id" element={<PlanScreen mode="view" data={data} reload={reload} stale={stale} />} />
        <Route path="/plans/:id/edit" element={<PlanScreen mode="edit" data={data} reload={reload} stale={stale} />} />
        <Route path="/plans/:id/copy" element={<PlanScreen mode="copy" data={data} reload={reload} stale={stale} />} />
        <Route path="/products" element={<Products products={data.products} />} />
        <Route path="/products/new" element={<EditProduct reload={reload} />} />
        <Route path="/products/:id/edit" element={<ProductScreen data={data} reload={reload} stale={stale} />} />
        <Route path="*" element={<MissingPage />} />
      </Routes>}
    </main>
  </Shell>;
}

const container = document.querySelector('#root');
if (container) createRoot(container).render(<BrowserRouter><App /></BrowserRouter>);
