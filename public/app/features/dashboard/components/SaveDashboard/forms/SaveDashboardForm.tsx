import { css } from '@emotion/css';
import { useMemo, useState } from 'react';

import { GrafanaTheme2 } from '@grafana/data';
import { selectors } from '@grafana/e2e-selectors';
import { Trans, t } from '@grafana/i18n';
import { config } from '@grafana/runtime';
import { Dashboard } from '@grafana/schema';
import { Button, Checkbox, TextArea, useStyles2, Stack } from '@grafana/ui';
import { DashboardModel } from 'app/features/dashboard/state/DashboardModel';
import { SaveDashboardResponseDTO } from 'app/types/dashboard';

import { GenAIDashboardChangesButton } from '../../GenAI/GenAIDashboardChangesButton';
import { SaveDashboardData, SaveDashboardOptions } from '../types';

export type SaveProps = {
  dashboard: DashboardModel; // original
  isLoading: boolean;
  saveModel: SaveDashboardData; // already cloned
  onCancel: () => void;
  onSuccess: () => void;
  onSubmit?: (
    saveModel: Dashboard,
    options: SaveDashboardOptions,
    dashboard: DashboardModel
  ) => Promise<SaveDashboardResponseDTO>;
  options: SaveDashboardOptions;
  onOptionsChange: (opts: SaveDashboardOptions) => void;
};

export const SaveDashboardForm = ({
  dashboard,
  isLoading,
  saveModel,
  options,
  onSubmit,
  onCancel,
  onSuccess,
  onOptionsChange,
}: SaveProps) => {
  const hasTimeChanged = useMemo(() => dashboard.hasTimeChanged(), [dashboard]);
  const hasVariableChanged = useMemo(() => dashboard.hasVariablesChanged(), [dashboard]);

  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState(options.message);
  const styles = useStyles2(getStyles);

  return (
    <form
      onSubmit={async (event) => {
        event.preventDefault();
        if (!onSubmit) {
          return;
        }
        setSaving(true);
        options = { ...options, message };
        const result = await onSubmit(saveModel.clone, options, dashboard);
        if (result.status === 'success') {
          onSuccess();
        } else {
          setSaving(false);
        }
      }}
      style={{ maxWidth: 600 }}
    >
      <Stack gap={2} direction="column" alignItems="flex-start">
        {hasTimeChanged && (
          <Checkbox
            checked={!!options.saveTimerange}
            onChange={() =>
              onOptionsChange({
                ...options,
                saveTimerange: !options.saveTimerange,
              })
            }
            label={t(
              'dashboard.save-dashboard-form.label-current-range-dashboard-default',
              'Сохранить текущий временной диапазон по умолчанию'
            )}
            aria-label={selectors.pages.SaveDashboardModal.saveTimerange}
          />
        )}
        {hasVariableChanged && (
          <Checkbox
            checked={!!options.saveVariables}
            onChange={() =>
              onOptionsChange({
                ...options,
                saveVariables: !options.saveVariables,
              })
            }
            label={t(
              'dashboard.save-dashboard-form.label-current-variable-values-dashboard-default',
              'Сохранить текущие переменные по умолчанию'
            )}
            aria-label={selectors.pages.SaveDashboardModal.saveVariables}
          />
        )}
        <div className={styles.message}>
          {config.featureToggles.aiGeneratedDashboardChanges && (
            <GenAIDashboardChangesButton
              dashboard={dashboard}
              onGenerate={(text) => {
                onOptionsChange({
                  ...options,
                  message: text,
                });
                setMessage(text);
              }}
              disabled={!saveModel.hasChanges}
            />
          )}
          <TextArea
            value={message}
            onChange={(e) => {
              onOptionsChange({
                ...options,
                message: e.currentTarget.value,
              });
              setMessage(e.currentTarget.value);
            }}
            placeholder={t(
              'dashboard.save-dashboard-form.placeholder-describe-changes',
              'Добавьте описание изменений.'
            )}
            autoFocus
            rows={5}
          />
        </div>

        <Stack alignItems="center">
          <Button variant="secondary" onClick={onCancel} fill="outline">
            <Trans i18nKey="dashboard.save-dashboard-form.cancel">Отмена</Trans>
          </Button>
          <Button
            type="submit"
            disabled={!saveModel.hasChanges || isLoading}
            icon={saving ? 'spinner' : undefined}
            aria-label={selectors.pages.SaveDashboardModal.save}
          >
            {isLoading
              ? t('dashboard.save-dashboard-form.saving', 'Сохранение...')
              : t('dashboard.save-dashboard-form.save', 'Сохранить')}
          </Button>
          {!saveModel.hasChanges && (
            <div>
              <Trans i18nKey="dashboard.save-dashboard-form.no-changes-to-save">Нет изменений для сохранения</Trans>
            </div>
          )}
        </Stack>
      </Stack>
    </form>
  );
};

function getStyles(theme: GrafanaTheme2) {
  return {
    message: css({
      display: 'flex',
      alignItems: 'end',
      flexDirection: 'column',
      width: '100%',
    }),
  };
}
