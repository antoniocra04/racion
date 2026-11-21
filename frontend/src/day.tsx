import { useState } from 'react';
import { Link, useNavigate } from 'react-router';
import { Button, Icon, PageHeading, buttonClass } from '@home-apps/ui';
import { decimal, grams, kcal, request, type Nutrition, type Plan, type Reload } from './api';

const macroNames = [['protein', 'Белки'], ['fat', 'Жиры'], ['carbs', 'Углеводы']] as const;

function compactNutrition({ kcal: energy, protein, fat, carbs }: Nutrition) {
  return `${kcal(energy)} ккал · Б ${grams(protein)} · Ж ${grams(fat)} · У ${grams(carbs)}`;
}

function Macros({ nutrition }: { nutrition: Nutrition }) {
  return <dl className="day-macros">
    {macroNames.map(([key, label]) => <div key={key}>
      <dt><i className={'macro-dot macro-dot--' + key} aria-hidden="true" />{label}</dt>
      <dd className="ui-number">{grams(nutrition[key])} г</dd>
    </div>)}
  </dl>;
}

export function Day({ plan, plans, reload }: { plan?: Plan; plans: Plan[]; reload: Reload }) {
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  if (!plan) return <section className="empty">
    <PageHeading>Рацион на день</PageHeading>
    <p className="ui-muted">Добавь приёмы пищи и граммовки продуктов.</p>
    <Link className={buttonClass('primary')} to="/plans/new"><Icon name="plus" />Создать рацион</Link>
  </section>;

  async function activate(id: number) {
    setPending(true);
    setError('');
    try {
      await request('/plans/' + id + '/activate', { method: 'POST' });
      await reload();
      navigate('/');
    } catch (error) {
      setError((error as Error).message);
    } finally {
      setPending(false);
    }
  }

  const actions = plan.active
    ? <label className="plan-switch">
        <select aria-label="Активный рацион" value={plan.id} onChange={(event) => void activate(Number(event.currentTarget.value))} disabled={pending}>
          {plans.map((option) => <option key={option.id} value={option.id}>{option.name}</option>)}
        </select>
        <Icon name="chevron" />
      </label>
    : <Button kind="primary" onClick={() => void activate(plan.id)} disabled={pending}>Сделать активным</Button>;

  return <>
    <PageHeading actions={actions}>{plan.active ? 'Рацион на день' : plan.name}</PageHeading>
    {error && <p className="ui-error page-error" role="alert">{error}</p>}
    <section className="nutrition" aria-labelledby="day-total">
      <div className="day-energy">
        <h2 id="day-total">Итого за день</h2>
        <p className="day-kcal ui-number">{kcal(plan.nutrition.kcal)} <span>ккал</span></p>
      </div>
      <Macros nutrition={plan.nutrition} />
    </section>
    <section className="meals" aria-labelledby="meals-title">
      <h2 id="meals-title">Приёмы пищи</h2>
      {plan.meals.length ? <div className="meal-list">
        {plan.meals.map((meal, index) => <details className="meal" key={index}>
          <summary><span>{meal.name}</span><span className="meal-kcal ui-number">{kcal(meal.nutrition.kcal)} ккал</span><Icon name="chevron" /></summary>
          <div className="meal-content">
            <p className="meal-nutrition ui-number">
              {macroNames.map(([key, label]) => <span key={key}>{label} {grams(meal.nutrition[key])} г</span>)}
            </p>
            {meal.items.length ? <ul className="ingredients">
              {meal.items.map((item, position) => <li key={position}>
                <div><strong>{item.productName}</strong><small>
                  {item.productState && <>{item.productState} · </>}{compactNutrition(item.nutrition)}
                </small></div>
                <span className="ingredient-weight ui-number">{decimal(item.grams)} г</span>
              </li>)}
            </ul> : <p className="ui-muted">В этом приёме пока нет продуктов.</p>}
          </div>
        </details>)}
      </div> : <p className="ui-muted">Добавь приёмы пищи в редакторе рациона.</p>}
    </section>
  </>;
}

export function Plans({ plans, reload }: { plans: Plan[]; reload: Reload }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  async function action(plan: Plan, intent: 'activate' | 'delete') {
    if (intent === 'delete' && !window.confirm('Удалить рацион «' + plan.name + '»?')) return;
    setBusy(true);
    setError('');
    try {
      await request('/plans/' + plan.id + (intent === 'activate' ? '/activate' : ''), {
        method: intent === 'activate' ? 'POST' : 'DELETE',
      });
      await reload();
      if (intent === 'activate') navigate('/');
    } catch (error) {
      setError((error as Error).message);
    } finally {
      setBusy(false);
    }
  }

  const add = <Link className={buttonClass('primary')} to="/plans/new"><Icon name="plus" />Создать рацион</Link>;
  return <>
    <PageHeading actions={add}>Версии рациона</PageHeading>
    {error && <p className="ui-error page-error" role="alert">{error}</p>}
    {plans.length ? <ul className="records">
      {plans.map((plan) => <li className="record" key={plan.id}>
        <div>
          <Link to={'/plans/' + plan.id}><strong>{plan.name}</strong></Link>
          <p className="ui-muted ui-number">{compactNutrition(plan.nutrition)}{plan.active && ' · Активный'}</p>
        </div>
        <div className="record-actions">
          {!plan.active && <Button kind="primary" disabled={busy} onClick={() => void action(plan, 'activate')}>Сделать активным</Button>}
          <Link className={buttonClass('quiet')} to={'/plans/' + plan.id + '/edit'}>Изменить</Link>
          <Link className={buttonClass('quiet')} to={'/plans/' + plan.id + '/copy'}>Копировать</Link>
          {(!plan.active || plans.length === 1) && <Button kind="danger" disabled={busy} onClick={() => void action(plan, 'delete')}>Удалить</Button>}
        </div>
      </li>)}
    </ul> : <p className="ui-muted">Пока нет сохранённых рационов.</p>}
  </>;
}
