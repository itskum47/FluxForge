import React, { createContext, useContext, useState, useEffect } from 'react';

interface TenantContextType {
    tenantID: string;
    setTenantID: (id: string) => void;
}

const TenantContext = createContext<TenantContextType | undefined>(undefined);

export function TenantProvider({ children }: { children: React.ReactNode }) {
    const [tenantID, setTenantID] = useState(() => {
        return localStorage.getItem('flux_tenant_id') || 'default';
    });

    useEffect(() => {
        localStorage.setItem('flux_tenant_id', tenantID);
    }, [tenantID]);

    return (
        <TenantContext.Provider value={{ tenantID, setTenantID }}>
            {children}
        </TenantContext.Provider>
    );
}

export function useTenant() {
    const context = useContext(TenantContext);
    if (context === undefined) {
        throw new Error('useTenant must be used within a TenantProvider');
    }
    return context;
}
