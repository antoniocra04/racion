import { useRef, useState } from 'react';
import { useNavigate } from 'react-router';
import { Button, Icon, PageHeading } from '@home-apps/ui';
import { ApiError, decimal, request, type Plan, type Product, type Reload } from './api';
import { AddProductDialog, type ProductOption } from './products';

interface DraftItem { key: number; productId: number; grams: string }
interface DraftMeal { key: number; name: string; items: DraftItem[] }
type Target = { meal: number; item?: number };
const maxMeals = 50;
const maxItems = 100;

function initialMeals(plan?: Plan): DraftMeal[] {
  if (!plan) return [{ key: 0, name: 'Завтрак', items: [] }];
  let key = 0;
  return plan.meals.map((meal) => ({
    key: key++,
    name: meal.name,
    items: meal.items.map((item) => ({ key: key++, productId: item.productId, grams: decimal(item.grams) })),
  }));
}

export function Editor({ plan, mode, products, reload }: {
  plan?: Plan;
  mode: 'new' | 'edit' | 'copy';
  products: Product[];
  reload: Reload;
}) {
  const [name, setName] = useState(() => mode === 'copy' && plan
    ? Array.from(plan.name).slice(0, 92).join('') + ' — копия'
    : plan?.name || '');
  const [meals, setMeals] = useState(() => initialMeals(plan));
  const [target, setTarget] = useState<Target | null>(null);
  const [savedProducts, setSavedProducts] = useState<ProductOption[]>([]);
  const [error, setError] = useState('');
  const [fields, setFields] = useState<Record<string, string>>({});
  const [pending, setPending] = useState(false);
  const nextKey = useRef(1 + meals.reduce((sum, meal) => sum + 1 + meal.items.length, 0));
  const saveLock = useRef(false);
  const navigate = useNavigate();
  const [edit] = useState(() => mode === 'edit' ? plan : undefined);
  const options = [...products, ...savedProducts.filter((saved) => !products.some(({ id }) => id === saved.id))];

  function setMeal(key: number, update: (meal: DraftMeal) => DraftMeal) {
    setFields({});
    setMeals((current) => current.map((meal) => meal.key === key ? update(meal) : meal));
  }

  function addItem(mealKey: number, productId = 0) {
    const item = { key: nextKey.current++, productId, grams: '' };
    setMeal(mealKey, (meal) => ({ ...meal, items: [...meal.items, item] }));
  }

  function updateItem(mealKey: number, key: number, values: Partial<Pick<DraftItem, 'productId' | 'grams'>>) {
    setMeal(mealKey, (meal) => ({
      ...meal,
      items: meal.items.map((item) => item.key === key ? { ...item, ...values } : item),
    }));
  }

  async function save() {
    if (saveLock.current) return;
    saveLock.current = true;
    setPending(true);
    setFields({});
    setError('');
    try {
      const { id } = await request<{ id: number }>(edit ? '/plans/' + edit.id : '/plans', {
        method: edit ? 'PUT' : 'POST',
        body: {
          name,
          revision: edit?.revision,
          meals: meals.map((meal) => ({
            name: meal.name,
            items: meal.items.map(({ productId, grams }) => ({ productId, grams })),
          })),
        },
      });
      await reload();
      navigate('/plans/' + id);
    } catch (error) {
      setError((error as Error).message);
      if (error instanceof ApiError) setFields(error.fields);
    } finally {
      saveLock.current = false;
      setPending(false);
    }
  }

  return <>
    <PageHeading>{mode === 'edit' ? 'Изменить рацион' : 'Создать рацион'}</PageHeading>
    <form className="fields plan-form" onSubmit={(event) => {
      event.preventDefault();
      void save();
    }}>
      {error && <p className="ui-error" role="alert">{error}</p>}
      <fieldset disabled={pending}>
        <label className="plan-name">Название рациона
          <input required maxLength={100} value={name} aria-invalid={!!fields.name} onChange={(event) => {
            setName(event.currentTarget.value);
            setFields((old) => ({ ...old, name: '' }));
          }} />
          {fields.name && <span className="ui-error">{fields.name}</span>}
        </label>
        <section className="edit-meals" aria-label="Приёмы пищи">
          {meals.map((meal, mi) => <section className="edit-meal" key={meal.key}>
            <div className="edit-meal-heading">
              <label>Название приёма пищи
                <input required maxLength={100} value={meal.name} aria-invalid={!!fields['meal.' + mi + '.name']} onChange={(event) => {
                  const value = event.currentTarget.value;
                  setMeal(meal.key, (current) => ({ ...current, name: value }));
                }} />
                {fields['meal.' + mi + '.name'] && <span className="ui-error">{fields['meal.' + mi + '.name']}</span>}
              </label>
              <Button kind="danger" aria-label={'Убрать приём ' + (meal.name || 'пищи')} onClick={() => {
                setFields({});
                setMeals((current) => current.filter((item) => item.key !== meal.key));
              }}><Icon name="trash" /></Button>
            </div>
            {meal.items.map((item, ii) => {
              const base = 'meal.' + mi + '.' + ii;
              return <div className="edit-item" key={item.key}>
                <label><span id={'product-label-' + item.key}>Продукт</span>
                  <select required value={item.productId || ''} aria-labelledby={'product-label-' + item.key} aria-invalid={!!fields[base + '.productId']}
                    onChange={(event) => {
                      const value = event.currentTarget.value;
                      if (value === 'new') setTarget({ meal: meal.key, item: item.key });
                      else updateItem(meal.key, item.key, { productId: Number(value) });
                    }}>
                    <option value="">Выбери из списка</option>
                    {options.map((product) => <option key={product.id} value={product.id}>{product.name}{product.state && ' · ' + product.state}</option>)}
                    <option value="new">Создать продукт…</option>
                  </select>
                  {fields[base + '.productId'] && <span className="ui-error">{fields[base + '.productId']}</span>}
                </label>
                <label>Вес, г
                  <input required inputMode="decimal" value={item.grams} aria-invalid={!!fields[base + '.grams']} onChange={(event) => updateItem(meal.key, item.key, { grams: event.currentTarget.value })} />
                  {fields[base + '.grams'] && <span className="ui-error">{fields[base + '.grams']}</span>}
                </label>
                <Button kind="quiet" aria-label="Убрать продукт" onClick={() => setMeal(meal.key, (current) => ({
                  ...current, items: current.items.filter((entry) => entry.key !== item.key),
                }))}><Icon name="close" /></Button>
              </div>;
            })}
            <div className="editor-add">
              {options.length > 0 && <Button kind="quiet" disabled={meal.items.length >= maxItems} onClick={() => addItem(meal.key)}><Icon name="plus" />Добавить продукт</Button>}
              <Button kind="quiet" disabled={meal.items.length >= maxItems} onClick={() => setTarget({ meal: meal.key })}>Создать продукт</Button>
            </div>
          </section>)}
          <Button disabled={meals.length >= maxMeals} onClick={() => {
            const meal = { key: nextKey.current++, name: '', items: [] };
            setMeals((current) => [...current, meal]);
          }}><Icon name="plus" />Добавить приём пищи</Button>
        </section>
      </fieldset>
      <div className="form-actions">
        <Button kind="quiet" disabled={pending} onClick={() => navigate(edit ? '/plans/' + edit.id : '/plans')}>Отмена</Button>
        <Button kind="primary" type="submit" disabled={pending}>{pending ? 'Сохраняю…' : 'Сохранить рацион'}</Button>
      </div>
    </form>
    {target && <AddProductDialog onClose={() => setTarget(null)} onSave={async (product) => {
      if (!await reload()) setSavedProducts((current) => [...current, product]);
      if (target.item === undefined) addItem(target.meal, product.id);
      else updateItem(target.meal, target.item, { productId: product.id });
      setTarget(null);
    }} />}
  </>;
}
