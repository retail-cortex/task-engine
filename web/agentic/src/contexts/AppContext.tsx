// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import React, { createContext, useContext, useState, useEffect } from 'react';
import { 
  ApiClient, 
  decodeOAuthTokenClaims, 
  SITE_ID, 
  ResponseError,
  BYPASS_USER_ID
} from '../api/client';

export interface AppContextType {
  theme: 'light' | 'dark';
  setTheme: React.Dispatch<React.SetStateAction<'light' | 'dark'>>;
  userToken: string | null;
  setUserToken: (token: string | null) => void;
  isAuthenticated: boolean;
  setIsAuthenticated: (val: boolean) => void;
  userName: string;
  setUserName: (name: string) => void;
  userEmail: string;
  setUserEmail: (email: string) => void;
  userPicture: string | null;
  setUserPicture: (pic: string | null) => void;
  userRole: string;
  setUserRole: (role: string) => void;
  userSites: any[];
  setUserSites: (sites: any[]) => void;
  activeSiteID: string;
  setActiveSiteID: (siteId: string) => void;
  allSites: any[];
  setAllSites: (sites: any[]) => void;
  allOrganizations: any[];
  setAllOrganizations: (orgs: any[]) => void;
  activeOrgID: string;
  setActiveOrgID: (orgId: string) => void;
  allUsers: any[];
  setAllUsers: (users: any[]) => void;
  activeUserId: string;
  setActiveUserId: (userId: string) => void;
  backendActive: boolean;
  setBackendActive: (val: boolean) => void;
  googleClientID: string;
  setGoogleClientID: (clientId: string) => void;
  schedulerLeader: boolean;
  setSchedulerLeader: (leader: boolean) => void;
  schedulerNodeID: string;
  setSchedulerNodeID: (nodeId: string) => void;
  schedulerTriggeredCount: number;
  setSchedulerTriggeredCount: React.Dispatch<React.SetStateAction<number>>;
  handleApiError: (err: any) => void;
  handleGoogleCredentialResponse: (response: any) => void;
  handleSignOut: () => void;
  syncProfileContext: () => void;
  syncSchedulerDiagnostics: () => void;
  handleSiteChange: (siteId: string) => void;
}

const AppContext = createContext<AppContextType | undefined>(undefined);

export const AppProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  // 0. Theme Swapping State Mechanism
  const [theme, setTheme] = useState<'light' | 'dark'>(
    () => (localStorage.getItem('theme') as 'light' | 'dark') || 'dark'
  );

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
  }, [theme]);

  // 1. Authentication & Google OAuth State
  const [userToken, setUserToken] = useState<string | null>(localStorage.getItem('oauth_token'));
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(!!localStorage.getItem('oauth_token'));
  const [userName, setUserName] = useState<string>(() => localStorage.getItem('oauth_name') || 'Hanna');
  const [userEmail, setUserEmail] = useState<string>(() => localStorage.getItem('oauth_email') || 'hanna@rmcguinness.altostrat.com');
  const [userPicture, setUserPicture] = useState<string | null>(() => localStorage.getItem('oauth_picture'));

  const [schedulerLeader, setSchedulerLeader] = useState<boolean>(true);
  const [schedulerNodeID, setSchedulerNodeID] = useState<string>('node-A');
  const [schedulerTriggeredCount, setSchedulerTriggeredCount] = useState<number>(0);
  const [backendActive, setBackendActive] = useState<boolean>(false);
  const [googleClientID, setGoogleClientID] = useState<string>("10781708810-t4ose5l4ck5hc9ouq7kk56dipq6a3h76.apps.googleusercontent.com");

  // Dynamic Role, Site mapping & multi-site switching context states
  const [userRole, setUserRole] = useState<string>("SITE_ASSOCIATE");
  const [userSites, setUserSites] = useState<any[]>([]);
  const [activeSiteID, setActiveSiteID] = useState<string>(
    () => localStorage.getItem('active_site_id') || SITE_ID
  );
  const [allSites, setAllSites] = useState<any[]>([]);
  const [allOrganizations, setAllOrganizations] = useState<any[]>([]);
  const [activeOrgID, setActiveOrgID] = useState<string>(
    () => localStorage.getItem('active_org_id') || 'ALL'
  );

  useEffect(() => {
    localStorage.setItem('active_org_id', activeOrgID);
  }, [activeOrgID]);
  const [allUsers, setAllUsers] = useState<any[]>([]);
  const [activeUserId, setActiveUserId] = useState<string>("");

  // Synchronise React state context dynamically with stateful ApiClient singleton session contexts!
  useEffect(() => {
    ApiClient.setToken(userToken);
  }, [userToken]);

  useEffect(() => {
    ApiClient.setActiveSiteId(activeSiteID);
  }, [activeSiteID]);

  useEffect(() => {
    ApiClient.setActiveUserId(activeUserId);
  }, [activeUserId]);

  // Centralised GORM API interceptor to handle Google OAuth token expirations (401 Unauthorized recovery pipeline!)
  const handleApiError = (err: any) => {
    console.error("[Operations API Error]", err);
    if (err instanceof ResponseError && err.status === 401) {
      console.warn("[OAuth Session] Credentials expired or validated incorrectly. Executing auto sign-out.");
      handleSignOut();
    }
  };

  const handleGoogleCredentialResponse = (response: any) => {
    const idToken = response.credential;
    
    const claims = decodeOAuthTokenClaims(idToken);
    if (claims) {
      setUserName(claims.name);
      setUserEmail(claims.email);
      setUserPicture(claims.picture);

      localStorage.setItem('oauth_name', claims.name);
      localStorage.setItem('oauth_email', claims.email);
      if (claims.picture) {
        localStorage.setItem('oauth_picture', claims.picture);
      } else {
        localStorage.removeItem('oauth_picture');
      }
    }

    localStorage.setItem('oauth_token', idToken);
    setUserToken(idToken);
    setIsAuthenticated(true);
  };

  const handleSignOut = () => {
    localStorage.removeItem('oauth_token');
    localStorage.removeItem('oauth_name');
    localStorage.removeItem('oauth_email');
    localStorage.removeItem('oauth_picture');
    
    // Purge the A2UI session ID to force a brand-new backend session on next login!
    sessionStorage.removeItem('a2ui_session_id');

    setUserToken(null);
    setIsAuthenticated(false);
    setUserPicture(null);
    setUserName('Hanna');
    setUserEmail('hanna@rmcguinness.altostrat.com');
    setUserRole("SITE_ASSOCIATE");
    setUserSites([]);
    setAllSites([]);
    setAllUsers([]);
    setAllOrganizations([]);
    setActiveOrgID('ALL');
    localStorage.removeItem('active_org_id');
    setActiveUserId(BYPASS_USER_ID);
  };

  const syncProfileContext = () => {
    ApiClient.fetchUserProfile()
      .then((user: any) => {
        const userId = user.Email || user.ID || user.id;
        setActiveUserId(userId);
        if (user.Name) setUserName(user.Name);
        if (user.Email) setUserEmail(user.Email);

        // Recover dynamic roles list preloaded from GORM
        const roleNames = user.Roles ? user.Roles.map((r: any) => r.Name) : [];
        let activeRole = "SITE_ASSOCIATE";
        if (roleNames.includes("ADMIN")) {
          activeRole = "ADMIN";
        } else if (roleNames.includes("REGION_MANAGER")) {
          activeRole = "REGION_MANAGER";
        } else if (roleNames.includes("SITE_MANAGER")) {
          activeRole = "SITE_MANAGER";
        } else if (roleNames.includes("SITE_3P")) {
          activeRole = "SITE_3P";
        }
        setUserRole(activeRole);

        // Recover preloaded physical store sites mappings
        const sitesList = user.Sites || [];
        setUserSites(sitesList);

        // Resolve context site target
        let initialSite = activeSiteID;
        const isAdminOrRegionManager = activeRole === "ADMIN" || activeRole === "REGION_MANAGER";
        if (sitesList.length > 0 && !isAdminOrRegionManager) {
          const hasActive = sitesList.some((s: any) => s.id === activeSiteID || s.ID === activeSiteID);
          if (!hasActive) {
            initialSite = sitesList[0].id || sitesList[0].ID;
            setActiveSiteID(initialSite);
            localStorage.setItem('active_site_id', initialSite);
          }
        }

        // Fetch master list datasets globally based on authorization profile scopes
        if (activeRole === "ADMIN" || activeRole === "REGION_MANAGER") {
          ApiClient.fetchSites()
            .then(data => setAllSites(data))
            .catch(console.error);

          ApiClient.fetchOrganizations()
            .then(data => setAllOrganizations(data || []))
            .catch(console.error);
        }

        if (activeRole === "ADMIN" || activeRole === "REGION_MANAGER" || activeRole === "SITE_MANAGER") {
          ApiClient.fetchUsers()
            .then(data => setAllUsers(data))
            .catch(console.error);
        }
      })
      .catch((err) => {
        handleApiError(err);
      });
  };

  const syncSchedulerDiagnostics = () => {
    ApiClient.fetchSchedulerStatus()
      .then((status: any) => {
        setSchedulerLeader(status.is_leader);
        setSchedulerNodeID(status.node_id);
      })
      .catch(handleApiError);
  };

  const handleSiteChange = (siteId: string) => {
    setActiveSiteID(siteId);
    localStorage.setItem('active_site_id', siteId);
  };

  // 3. Fetch and Sync Pipeline Hooks (Readiness probe is completely public and unauthenticated!)
  useEffect(() => {
    ApiClient.probeReadiness()
      .then((data) => {
        setBackendActive(true);
        if (data && data.client_id) {
          setGoogleClientID(data.client_id);
        }
        if (isAuthenticated) {
          syncProfileContext();
          syncSchedulerDiagnostics();
        }
      })
      .catch(() => {
        setBackendActive(false);
      });
  }, [isAuthenticated, userToken]);

  return (
    <AppContext.Provider value={{
      theme,
      setTheme,
      userToken,
      setUserToken,
      isAuthenticated,
      setIsAuthenticated,
      userName,
      setUserName,
      userEmail,
      setUserEmail,
      userPicture,
      setUserPicture,
      userRole,
      setUserRole,
      userSites,
      setUserSites,
      activeSiteID,
      setActiveSiteID,
      allSites,
      setAllSites,
      allOrganizations,
      setAllOrganizations,
      activeOrgID,
      setActiveOrgID,
      allUsers,
      setAllUsers,
      activeUserId,
      setActiveUserId,
      backendActive,
      setBackendActive,
      googleClientID,
      setGoogleClientID,
      schedulerLeader,
      setSchedulerLeader,
      schedulerNodeID,
      setSchedulerNodeID,
      schedulerTriggeredCount,
      setSchedulerTriggeredCount,
      handleApiError,
      handleGoogleCredentialResponse,
      handleSignOut,
      syncProfileContext,
      syncSchedulerDiagnostics,
      handleSiteChange,
    }}>
      {children}
    </AppContext.Provider>
  );
};

export const useAppContext = () => {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error('useAppContext must be used within an AppProvider');
  }
  return context;
};
