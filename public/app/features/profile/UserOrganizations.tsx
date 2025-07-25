import { PureComponent } from 'react';

import { selectors } from '@grafana/e2e-selectors';
import { Trans, t } from '@grafana/i18n';
import { Button, LoadingPlaceholder, ScrollContainer } from '@grafana/ui';
import { UserDTO, UserOrg } from 'app/types/user';

export interface Props {
  user: UserDTO | null;
  orgs: UserOrg[];
  isLoading: boolean;
  setUserOrg: (org: UserOrg) => void;
}

export class UserOrganizations extends PureComponent<Props> {
  render() {
    const { isLoading, orgs, user } = this.props;

    if (isLoading) {
      return (
        <LoadingPlaceholder
          text={t('profile.user-organizations.text-loading-organizations', 'Loading organizations...')}
        />
      );
    }

    if (orgs.length === 0) {
      return null;
    }

    return (
      <div>
        <h3 className="page-sub-heading">
          <Trans i18nKey="user-orgs.title">Organizations</Trans>
        </h3>

        <ScrollContainer overflowY="visible" overflowX="auto" width="100%">
          <table className="filter-table form-inline" data-testid={selectors.components.UserProfile.orgsTable}>
            <thead>
              <tr>
                <th>
                  <Trans i18nKey="user-orgs.name-column">Name</Trans>
                </th>
                <th>
                  <Trans i18nKey="user-orgs.role-column">Role</Trans>
                </th>
                <th />
              </tr>
            </thead>
            <tbody>
              {orgs.map((org: UserOrg, index) => {
                return (
                  <tr key={index}>
                    <td>{org.name}</td>
                    <td>{org.role}</td>
                    <td className="text-right">
                      {org.orgId === user?.orgId ? (
                        <Button variant="secondary" size="sm" disabled>
                          <Trans i18nKey="user-orgs.current-org-button">Current</Trans>
                        </Button>
                      ) : (
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => {
                            this.props.setUserOrg(org);
                          }}
                        >
                          <Trans i18nKey="user-orgs.select-org-button">Select organisation</Trans>
                        </Button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </ScrollContainer>
      </div>
    );
  }
}

export default UserOrganizations;
