import { TimeOption } from '@grafana/data';
import { t } from '@grafana/i18n';

import { ComboboxOption } from '../Combobox/types';

export const getQuickOptions: () => TimeOption[] = () => [
  { from: 'now-5m', to: 'now', display: t('grafana-ui.date-time-pickers.quick-options.last-5-mins', 'Последние 5 минут') },
  {
    from: 'now-15m',
    to: 'now',
    display: t('grafana-ui.date-time-pickers.quick-options.last-15-mins', 'Последние 15 минут'),
  },
  {
    from: 'now-30m',
    to: 'now',
    display: t('grafana-ui.date-time-pickers.quick-options.last-30-mins', 'Последние 30 минут'),
  },
  { from: 'now-1h', to: 'now', display: t('grafana-ui.date-time-pickers.quick-options.last-1-hour', 'Последний 1 час') },
  { from: 'now-3h', to: 'now', display: t('grafana-ui.date-time-pickers.quick-options.last-3-hours', 'Последние 3 часа') },
  { from: 'now-6h', to: 'now', display: t('grafana-ui.date-time-pickers.quick-options.last-6-hours', 'Последние 6 часов') },
  {
    from: 'now-12h',
    to: 'now',
    display: t('grafana-ui.date-time-pickers.quick-options.last-12-hours', 'Последние 12 часов'),
  },
  {
    from: 'now-24h',
    to: 'now',
    display: t('grafana-ui.date-time-pickers.quick-options.last-24-hours', 'Последние 24 часа'),
  },
  { from: 'now-2d', to: 'now', display: t('grafana-ui.date-time-pickers.quick-options.last-2-days', 'Последние 2 дня') },
  { from: 'now-7d', to: 'now', display: t('grafana-ui.date-time-pickers.quick-options.last-7-days', 'Последние 7 дней') },
  { from: 'now-30d', to: 'now', display: t('grafana-ui.date-time-pickers.quick-options.last-30-days', 'Последние 30 дней') },
  { from: 'now-90d', to: 'now', display: t('grafana-ui.date-time-pickers.quick-options.last-90-days', 'Последние 90 дней') },
  {
    from: 'now-6M',
    to: 'now',
    display: t('grafana-ui.date-time-pickers.quick-options.last-6-months', 'Последние 6 месяцев'),
  },
  { from: 'now-1y', to: 'now', display: t('grafana-ui.date-time-pickers.quick-options.last-1-year', 'Последний 1 год') },
  { from: 'now-1d/d+7h', to: 'now/1d-1d+7h', display: 'Сводка с 7 до 7 утра' },
  // { from: 'now-2y', to: 'now', display: t('grafana-ui.date-time-pickers.quick-options.last-2-years', 'Last 2 years') },
  // { from: 'now-5y', to: 'now', display: t('grafana-ui.date-time-pickers.quick-options.last-5-years', 'Last 5 years') },
  { from: 'now-1d/d', to: 'now-1d/d', display: t('grafana-ui.date-time-pickers.quick-options.yesterday', 'Вчера') },
  {
    from: 'now-2d/d',
    to: 'now-2d/d',
    display: t('grafana-ui.date-time-pickers.quick-options.day-before-yesterday', 'Позавчера'),
  },
  {
    from: 'now-7d/d',
    to: 'now-7d/d',
    display: t('grafana-ui.date-time-pickers.quick-options.this-day-last-week', 'В этот день на прошлой неделе'),
  },
  {
    from: 'now-1w/w',
    to: 'now-1w/w',
    display: t('grafana-ui.date-time-pickers.quick-options.previous-week', 'Предыдущая неделя'),
  },
  {
    from: 'now-1M/M',
    to: 'now-1M/M',
    display: t('grafana-ui.date-time-pickers.quick-options.previous-month', 'Предыдущий месяц'),
  },
  // {
  //   from: 'now-1Q/fQ',
  //   to: 'now-1Q/fQ',
  //   display: t('grafana-ui.date-time-pickers.quick-options.previous-fiscal-quarter', 'Previous fiscal quarter'),
  // },
  {
    from: 'now-1y/y',
    to: 'now-1y/y',
    display: t('grafana-ui.date-time-pickers.quick-options.previous-year', 'Предыдущий год'),
  },
  // {
  //   from: 'now-1y/fy',
  //   to: 'now-1y/fy',
  //   display: t('grafana-ui.date-time-pickers.quick-options.previous-fiscal-year', 'Previous fiscal year'),
  // },
  { from: 'now/d', to: 'now/d', display: t('grafana-ui.date-time-pickers.quick-options.today', 'Сегодня') },
  // { from: 'now/d', to: 'now', display: t('grafana-ui.date-time-pickers.quick-options.today-so-far', 'Today so far') },
  { from: 'now/w', to: 'now/w', display: t('grafana-ui.date-time-pickers.quick-options.this-week', 'Текущая неделя') },
  // {
  //   from: 'now/w',
  //   to: 'now',
  //   display: t('grafana-ui.date-time-pickers.quick-options.this-week-so-far', 'This week so far'),
  // },
  { from: 'now/M', to: 'now/M', display: t('grafana-ui.date-time-pickers.quick-options.this-month', 'Текущий месяц') },
  // {
  //   from: 'now/M',
  //   to: 'now',
  //   display: t('grafana-ui.date-time-pickers.quick-options.this-month-so-far', 'This month so far'),
  // },
  { from: 'now/y', to: 'now/y', display: t('grafana-ui.date-time-pickers.quick-options.this-year', 'Текущий год') },
  // {
  //   from: 'now/y',
  //   to: 'now',
  //   display: t('grafana-ui.date-time-pickers.quick-options.this-year-so-far', 'This year so far'),
  // },
  // {
  //   from: 'now/fQ',
  //   to: 'now',
  //   display: t('grafana-ui.date-time-pickers.quick-options.this-fiscal-quarter-so-far', 'This fiscal quarter so far'),
  // },
  // {
  //   from: 'now/fQ',
  //   to: 'now/fQ',
  //   display: t('grafana-ui.date-time-pickers.quick-options.this-fiscal-quarter', 'This fiscal quarter'),
  // },
  // {
  //   from: 'now/fy',
  //   to: 'now',
  //   display: t('grafana-ui.date-time-pickers.quick-options.this-fiscal-year-so-far', 'This fiscal year so far'),
  // },
  // {
  //   from: 'now/fy',
  //   to: 'now/fy',
  //   display: t('grafana-ui.date-time-pickers.quick-options.this-fiscal-year', 'This fiscal year'),
  // },
];

export const getMonthOptions: () => Array<ComboboxOption<number>> = () => [
  { label: t('grafana-ui.date-time-pickers.month-options.label-january', 'Января'), value: 0 },
  { label: t('grafana-ui.date-time-pickers.month-options.label-february', 'Февраль'), value: 1 },
  { label: t('grafana-ui.date-time-pickers.month-options.label-march', 'Март'), value: 2 },
  { label: t('grafana-ui.date-time-pickers.month-options.label-april', 'Апрель'), value: 3 },
  { label: t('grafana-ui.date-time-pickers.month-options.label-may', 'Май'), value: 4 },
  { label: t('grafana-ui.date-time-pickers.month-options.label-june', 'Июнь'), value: 5 },
  { label: t('grafana-ui.date-time-pickers.month-options.label-july', 'Июль'), value: 6 },
  { label: t('grafana-ui.date-time-pickers.month-options.label-august', 'Август'), value: 7 },
  { label: t('grafana-ui.date-time-pickers.month-options.label-september', 'Сентябрь'), value: 8 },
  { label: t('grafana-ui.date-time-pickers.month-options.label-october', 'Октябрь'), value: 9 },
  { label: t('grafana-ui.date-time-pickers.month-options.label-november', 'Ноябрь'), value: 10 },
  { label: t('grafana-ui.date-time-pickers.month-options.label-december', 'Декабрь'), value: 11 },
];
