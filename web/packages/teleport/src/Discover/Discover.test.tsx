/**
 * Teleport
 * Copyright (C) 2023  Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import { MemoryRouter } from 'react-router';

import { render, screen } from 'design/utils/testing';
import { Resource } from 'gen-proto-ts/teleport/userpreferences/v1/onboard_pb';
import { InfoGuidePanelProvider } from 'shared/components/SlidingSidePanel/InfoGuide';

import cfg from 'teleport/config';
import { Discover, DiscoverComponent } from 'teleport/Discover/Discover';
import { ResourceViewConfig } from 'teleport/Discover/flow';
import {
  APPLICATIONS,
  DATABASES,
  DATABASES_UNGUIDED,
  DATABASES_UNGUIDED_DOC,
  KUBERNETES,
  SERVERS,
} from 'teleport/Discover/SelectResource/resources';
import { getOSSFeatures } from 'teleport/features';
import { FeaturesContextProvider } from 'teleport/FeaturesContext';
import { createTeleportContext, getAcl } from 'teleport/mocks/contexts';
import { makeDefaultUserPreferences } from 'teleport/services/userPreferences/userPreferences';
import TeleportContextProvider from 'teleport/TeleportContextProvider';
import { makeTestUserContext } from 'teleport/User/testHelpers/makeTestUserContext';
import { mockUserContextProviderWith } from 'teleport/User/testHelpers/mockUserContextWith';

import {
  resourceSpecConnectMyComputer,
  resourceSpecSamlGcp,
} from './Fixtures/fixtures';
import { ResourceKind } from './Shared';
import { getGuideTileId } from './testUtils';
import { DiscoverUpdateProps, useDiscover } from './useDiscover';

beforeEach(() => {
  jest.restoreAllMocks();
});

type createProps = {
  initialEntry?: string;
  preferredResource?: Resource;
};

const create = ({ initialEntry = '', preferredResource }: createProps) => {
  jest.spyOn(window.navigator, 'userAgent', 'get').mockReturnValue('Macintosh');

  const defaultPref = makeDefaultUserPreferences();
  defaultPref.onboard.preferredResources = preferredResource
    ? [preferredResource]
    : [];

  mockUserContextProviderWith(
    makeTestUserContext({ preferences: defaultPref })
  );

  const userAcl = getAcl();
  const ctx = createTeleportContext({ customAcl: userAcl });

  return render(
    <MemoryRouter
      initialEntries={[
        { pathname: cfg.routes.discover, state: { entity: initialEntry } },
      ]}
    >
      <TeleportContextProvider ctx={ctx}>
        <FeaturesContextProvider value={getOSSFeatures()}>
          <InfoGuidePanelProvider>
            <Discover />
          </InfoGuidePanelProvider>
        </FeaturesContextProvider>
      </TeleportContextProvider>
    </MemoryRouter>
  );
};

test('displays all resources by default', () => {
  create({});

  expect(
    screen
      .getAllByTestId(getGuideTileId({ kind: ResourceKind.Server }))
      .concat(
        screen.getAllByTestId(
          getGuideTileId({ kind: ResourceKind.ConnectMyComputer })
        )
      )
  ).toHaveLength(SERVERS.length);
  expect(
    screen.getAllByTestId(getGuideTileId({ kind: ResourceKind.Database }))
  ).toHaveLength(
    DATABASES.length + DATABASES_UNGUIDED.length + DATABASES_UNGUIDED_DOC.length
  );
  expect(
    screen.getAllByTestId(getGuideTileId({ kind: ResourceKind.Application }))
  ).toHaveLength(APPLICATIONS.length);
  expect(
    screen.getAllByTestId(getGuideTileId({ kind: ResourceKind.Kubernetes }))
  ).toHaveLength(KUBERNETES.length);
});

test('location state applies filter/search', () => {
  create({
    initialEntry: 'desktop',
    preferredResource: Resource.WEB_APPLICATIONS,
  });

  expect(
    screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Application }))
  ).not.toBeInTheDocument();
  expect(
    screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Server }))
  ).not.toBeInTheDocument();
  expect(
    screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Database }))
  ).not.toBeInTheDocument();
  expect(
    screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Kubernetes }))
  ).not.toBeInTheDocument();
});

describe('location state', () => {
  test('displays servers when the location state is server', () => {
    create({ initialEntry: 'server' });

    expect(
      screen
        .getAllByTestId(getGuideTileId({ kind: ResourceKind.Server }))
        .concat(
          screen.getAllByTestId(
            getGuideTileId({ kind: ResourceKind.ConnectMyComputer })
          )
        )
    ).toHaveLength(SERVERS.length);

    // we assert three databases for servers because the naming convention includes "server"
    expect(
      screen.queryAllByTestId(getGuideTileId({ kind: ResourceKind.Database }))
    ).toHaveLength(4);

    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Desktop }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Application }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Kubernetes }))
    ).not.toBeInTheDocument();
  });

  test('displays desktops when the location state is desktop', () => {
    create({ initialEntry: 'desktop' });

    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Server }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Database }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Application }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Kubernetes }))
    ).not.toBeInTheDocument();
  });

  test('displays apps when the location state is application', () => {
    create({ initialEntry: 'application' });

    expect(
      screen.getAllByTestId(getGuideTileId({ kind: ResourceKind.Application }))
    ).toHaveLength(APPLICATIONS.length);

    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Server }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Desktop }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Database }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Kubernetes }))
    ).not.toBeInTheDocument();
  });

  test('displays databases when the location state is database', () => {
    create({ initialEntry: 'database' });

    expect(
      screen.getAllByTestId(getGuideTileId({ kind: ResourceKind.Database }))
    ).toHaveLength(
      DATABASES.length +
        DATABASES_UNGUIDED.length +
        DATABASES_UNGUIDED_DOC.length
    );

    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Server }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Desktop }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Application }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Kubernetes }))
    ).not.toBeInTheDocument();
  });

  test('displays kube resources when the location state is kubernetes', () => {
    create({ initialEntry: 'kubernetes' });

    expect(
      screen.getAllByTestId(getGuideTileId({ kind: ResourceKind.Kubernetes }))
    ).toHaveLength(KUBERNETES.length);

    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Server }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Desktop }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(getGuideTileId({ kind: ResourceKind.Database }))
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(ResourceKind.Application)
    ).not.toBeInTheDocument();
  });
});

const renderUpdate = (props: DiscoverUpdateProps) => {
  const defaultPref = makeDefaultUserPreferences();
  defaultPref.onboard.preferredResources = [Resource.WEB_APPLICATIONS];

  mockUserContextProviderWith(
    makeTestUserContext({ preferences: defaultPref })
  );

  const userAcl = getAcl();
  const ctx = createTeleportContext({ customAcl: userAcl });

  const MockComponent1 = () => {
    const { agentMeta } = useDiscover();
    return (
      <>
        {agentMeta.resourceName === 'saml2' ? agentMeta.resourceName : 'saml1'}
      </>
    );
  };

  const testViews: ResourceViewConfig[] = [
    {
      kind: ResourceKind.SamlApplication,
      views() {
        return [
          {
            title: 'MockComponent1',
            component: MockComponent1,
          },
        ];
      },
    },
  ];

  return render(
    <MemoryRouter
      initialEntries={[
        { pathname: cfg.routes.discover, state: { entity: '' } },
      ]}
    >
      <TeleportContextProvider ctx={ctx}>
        <InfoGuidePanelProvider>
          <DiscoverComponent eViewConfigs={testViews} updateFlow={props} />
        </InfoGuidePanelProvider>
      </TeleportContextProvider>
    </MemoryRouter>
  );
};

test('update flow: renders single component based on resourceSpec', () => {
  renderUpdate({
    resourceSpec: resourceSpecConnectMyComputer,
    agentMeta: { resourceName: '' },
  });

  expect(screen.queryByTestId(ResourceKind.Server)).not.toBeInTheDocument();

  expect(screen.queryByTestId(ResourceKind.Database)).not.toBeInTheDocument();

  expect(
    screen.queryByTestId(ResourceKind.Application)
  ).not.toBeInTheDocument();

  expect(screen.queryByTestId(ResourceKind.Kubernetes)).not.toBeInTheDocument();

  expect(screen.getByText('Sign In & Connect My Computer')).toBeInTheDocument();
});

test('update flow: agentMeta is prepopulated based on agentMeta', () => {
  renderUpdate({
    resourceSpec: resourceSpecSamlGcp,
    agentMeta: { resourceName: 'saml2' },
  });

  expect(screen.getByText('saml2')).toBeInTheDocument();
});
