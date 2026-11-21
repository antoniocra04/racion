import { useEffect, useId, useRef, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router';
import { Button, Icon, PageHeading, buttonClass } from '@home-apps/ui';
import { ApiError, creationToken, decimal, grams, kcal, request, type Product, type Reload } from './api';

type ProductValues = Pick<Product, 'name' | 'state'> & { protein: string; fat: string; carbs: string };
export type ProductOption = Pick<Product, 'id' | 'name' | 'state'>;
const macros = [['protein', 'Белки, г', 4], ['fat', 'Жиры, г', 9], ['carbs', 'Углеводы, г', 4]] as const;

function initialProduct(product?: Product): ProductValues {
  return product
    ? { name: product.name, state: product.state, protein: decimal(product.per100.protein), fat: decimal(product.per100.fat), carbs: decimal(product.per100.carbs) }
    : { name: '', state: '', protein: '', fat: '', carbs: '' };
}

export function ProductForm({ product, onSave, onCancel }: {
  product?: Product;
  onSave: (product: ProductOption) => void | Promise<void>;
  onCancel: () => void;
}) {
  const [values, setValues] = useState(() => initialProduct(product));
  const [token] = useState(creationToken);
  const [error, setError] = useState('');
  const [fields, setFields] = useState<Record<string, string>>({});
  const [pending, setPending] = useState(false);
  const saveLock = useRef(false);
  const stateLabel = useId();
  const update = (key: keyof ProductValues, value: string) => {
    setValues((old) => ({ ...old, [key]: value }));
    setFields((old) => ({ ...old, [key]: '' }));
  };

  const allMacrosValid = macros.every(([key]) => /^\d+(?:[.,]\d*)?$|^[.,]\d+$/.test(values[key].trim()) && Number(values[key].replace(',', '.')) <= 1e6);
  const energy = allMacrosValid ? macros.reduce((sum, [key, , ratio]) => sum + Number(values[key].replace(',', '.')) * ratio, 0) : null;

  async function save() {
    if (saveLock.current) return;
    saveLock.current = true;
    setPending(true);
    setError('');
    setFields({});
    try {
      const { id } = await request<{ id: number }>(product ? '/products/' + product.id : '/products', {
        method: product ? 'PUT' : 'POST',
        body: { ...values, revision: product?.revision, token },
      });
      await onSave({ id, name: values.name.trim(), state: values.state });
    } catch (error) {
      setError((error as Error).message);
      if (error instanceof ApiError) setFields(error.fields);
    } finally {
      saveLock.current = false;
      setPending(false);
    }
  }

  return <form className="fields product-form" onSubmit={(event) => {
    event.preventDefault();
    void save();
  }}>
    {error && <p className="ui-error" role="alert">{error}</p>}
    <fieldset disabled={pending}>
      <label>Название
        <input required maxLength={100} value={values.name} aria-invalid={!!fields.name} onChange={(event) => update('name', event.currentTarget.value)} />
        {fields.name && <span className="ui-error">{fields.name}</span>}
      </label>
      <label><span id={stateLabel}>Состояние</span>
        <select value={values.state} aria-labelledby={stateLabel} aria-invalid={!!fields.state} onChange={(event) => update('state', event.currentTarget.value)}>
          <option value="">Не указано</option>
          <option value="сухой">Сухой</option><option value="сырой">Сырой</option><option value="готовый">Готовый</option>
        </select>
        {fields.state && <span className="ui-error">{fields.state}</span>}
      </label>
      <div className="form-section-label">На 100 г</div>
      <div className="macro-fields">
        {macros.map(([key, label]) => <label key={key}>{label}
          <input required inputMode="decimal" value={values[key]} aria-invalid={!!fields[key]} onChange={(event) => update(key, event.currentTarget.value)} />
          {fields[key] && <span className="ui-error">{fields[key]}</span>}
        </label>)}
      </div>
      <p className="energy-output ui-muted" aria-live="polite">Калорийность <output className="ui-number">{energy === null ? '—' : kcal(energy) + ' ккал'}</output></p>
    </fieldset>
    <div className="form-actions">
      <Button kind="quiet" onClick={onCancel} disabled={pending}>Отмена</Button>
      <Button type="submit" kind="primary" disabled={pending}>{pending ? 'Сохраняю…' : 'Сохранить'}</Button>
    </div>
  </form>;
}

export function AddProductDialog({ onSave, onClose }: { onSave: (product: ProductOption) => void | Promise<void>; onClose: () => void }) {
  const dialog = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const element = dialog.current;
    element?.showModal();
    return () => { if (element?.open) element.close(); };
  }, []);
  return <dialog ref={dialog} className="add-product-dialog" aria-labelledby="new-product-title" onCancel={(event) => {
    event.preventDefault();
    if (!dialog.current?.querySelector('fieldset:disabled')) onClose();
  }}>
    <h2 id="new-product-title">Добавить продукт</h2>
    <ProductForm onSave={onSave} onCancel={onClose} />
  </dialog>;
}

export function Products({ products }: { products: Product[] }) {
  const [params, setParams] = useSearchParams();
  const query = params.get('q') || '';
  const filtered = products.filter((product) => product.name.toLocaleLowerCase('ru').includes(query.trim().toLocaleLowerCase('ru')));
  return <>
    <PageHeading actions={<Link className={buttonClass('primary')} to="/products/new"><Icon name="plus" />Добавить продукт</Link>}>
      Мои продукты
    </PageHeading>
    {(products.length > 0 || query) && <label className="fields product-search">
      <span className="ui-sr-only">Поиск по названию</span>
      <input type="search" placeholder="Поиск по названию" value={query}
        onChange={(event) => setParams(event.currentTarget.value ? { q: event.currentTarget.value } : {}, { replace: true })} />
    </label>}
    {filtered.length ? <ul className="records">
      {filtered.map((product) => <li key={product.id}>
        <Link className="record product-record" to={'/products/' + product.id + '/edit'}>
          <div><strong>{product.name}</strong>{product.state && <p className="ui-muted">{product.state}</p>}</div>
          <div className="record-actions product-macros ui-number">
            <div><strong>{kcal(product.per100.kcal)} ккал</strong>
              <p className="ui-muted">Б {grams(product.per100.protein)} · Ж {grams(product.per100.fat)} · У {grams(product.per100.carbs)}</p></div>
            <Icon name="next" />
          </div>
        </Link>
      </li>)}
    </ul> : <p className="ui-muted">{query ? 'Ничего не найдено. Попробуй изменить запрос.' : 'Пока нет продуктов. Добавь продукт и его БЖУ.'}</p>}
  </>;
}

export function EditProduct({ product, reload }: { product?: Product; reload: Reload }) {
  const navigate = useNavigate();
  const [deleteError, setDeleteError] = useState('');
  const [deleting, setDeleting] = useState(false);
  async function remove() {
    if (!product || !window.confirm('Удалить продукт «' + product.name + '»?')) return;
    setDeleting(true);
    setDeleteError('');
    try {
      await request('/products/' + product.id, { method: 'DELETE' });
      await reload();
      navigate('/products');
    } catch (error) {
      setDeleteError((error as Error).message);
    } finally {
      setDeleting(false);
    }
  }

  return <>
    <PageHeading>{product ? 'Изменить продукт' : 'Добавить продукт'}</PageHeading>
    <ProductForm product={product} onCancel={() => navigate('/products')} onSave={async () => {
      await reload();
      navigate('/products');
    }} />
    {product && <div className="product-delete">
      {deleteError && <p className="ui-error page-error" role="alert">{deleteError}</p>}
      <Button kind="danger" disabled={deleting} onClick={() => void remove()}><Icon name="trash" />Удалить продукт</Button>
    </div>}
  </>;
}
