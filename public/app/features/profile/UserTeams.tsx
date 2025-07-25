import { PureComponent } from 'react';

import { Trans, t } from '@grafana/i18n';
import { LoadingPlaceholder, ScrollContainer } from '@grafana/ui';
import { Team } from 'app/types/teams';

export interface Props {
  teams: Team[];
  isLoading: boolean;
}

export class UserTeams extends PureComponent<Props> {
  render() {
    const { isLoading, teams } = this.props;

    if (isLoading) {
      return <LoadingPlaceholder text={t('profile.user-teams.text-loading-teams', 'Загрузка команд...')} />;
    }

    if (teams.length === 0) {
      return null;
    }

    return (
      <div>
        <h3 className="page-sub-heading">
          <Trans i18nKey="profile.user-teams.teams">Команды</Trans>
        </h3>
        <ScrollContainer overflowY="visible" overflowX="auto" width="100%">
          <table
            className="filter-table form-inline"
            aria-label={t('profile.user-teams.aria-label-user-teams-table', 'User teams table')}
          >
            <thead>
              <tr>
                <th />
                <th>
                  <Trans i18nKey="profile.user-teams.name">Name</Trans>
                </th>
                <th>
                  <Trans i18nKey="profile.user-teams.email">Email</Trans>
                </th>
                <th>
                  <Trans i18nKey="profile.user-teams.members">Пользователи</Trans>
                </th>
              </tr>
            </thead>
            <tbody>
              {teams.map((team: Team, index) => {
                return (
                  <tr key={index}>
                    <td className="width-4 text-center">
                      <img className="filter-table__avatar" src={team.avatarUrl} alt="" />
                    </td>
                    <td>{team.name}</td>
                    <td>{team.email}</td>
                    <td>{team.memberCount}</td>
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

export default UserTeams;
