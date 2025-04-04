import { css } from '@emotion/css';

import { Button, Modal } from '@grafana/ui';
import { t, Trans } from 'app/core/internationalization';

import { DashboardModel } from '../../state/DashboardModel';

import { SaveDashboardButton } from './SaveDashboardButton';

interface UnsavedChangesModalProps {
  dashboard: DashboardModel;
  onDiscard: () => void;
  onDismiss: () => void;
  onSaveSuccess?: () => void;
}

export const UnsavedChangesModal = ({ dashboard, onSaveSuccess, onDiscard, onDismiss }: UnsavedChangesModalProps) => {
  return (
    <Modal
      isOpen={true}
      title={t('dashboard.unsaved-changes-modal.title-unsaved-changes', 'Несохраненные изменения')}
      onDismiss={onDismiss}
      icon="exclamation-triangle"
      className={css({
        width: '500px',
      })}
    >
      <h5>
        <Trans i18nKey="dashboard.unsaved-changes-modal.changes">Вы хотите сохранить изменения?</Trans>
      </h5>
      <Modal.ButtonRow>
        <Button variant="secondary" onClick={onDismiss} fill="outline">
          <Trans i18nKey="dashboard.unsaved-changes-modal.cancel">Отмена</Trans>
        </Button>
        <Button variant="destructive" onClick={onDiscard}>
          <Trans i18nKey="dashboard.unsaved-changes-modal.discard">Отказаться</Trans>
        </Button>
        <SaveDashboardButton dashboard={dashboard} onSaveSuccess={onSaveSuccess} />
      </Modal.ButtonRow>
    </Modal>
  );
};
