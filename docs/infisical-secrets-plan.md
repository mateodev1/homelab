# Gestor de Secretos Tipo Infisical

## Objetivo

Crear un gestor de secretos por producto, subproyecto y entorno, inspirado en Infisical.

## Modelo

```text
Product
└── Project
    └── Environment
        └── Secret
```

Ejemplo:

```text
homelab
├── backend
│   ├── development
│   ├── staging
│   └── production
├── frontend
│   ├── development
│   └── production
└── cli
    ├── development
    └── production
```

## Funcionalidades

- Crear, editar, eliminar y buscar secretos.
- Ocultar valores por defecto.
- Revelar o copiar un secreto individual.
- Copiar todos los secretos del entorno seleccionado.
- Descargar todos los secretos como archivo `.env`.
- Exportar todos los secretos aunque la vista tenga filtros activos.
- Registrar auditoría de lecturas y exportaciones.
- Mantener historial de versiones.

## Exportación `.env`

Las acciones de copiar y descargar deben usar el mismo serializador.

```dotenv
DATABASE_URL="postgres://..."
API_KEY="..."
FEATURE_ENABLED="true"
```

El serializador debe soportar espacios, comillas, caracteres especiales y valores multilinea.

## API

```text
GET /products/{product}/projects/{project}/environments/{environment}/secrets
GET /products/{product}/projects/{project}/environments/{environment}/secrets/export
POST /products/{product}/projects/{project}/environments/{environment}/secrets
PUT /products/{product}/projects/{project}/environments/{environment}/secrets/{key}
DELETE /products/{product}/projects/{project}/environments/{environment}/secrets/{key}
```

El endpoint de exportación debe responder `text/plain` y aplicar los permisos del entorno.

## Seguridad

- Cifrar los valores en reposo.
- Nunca registrar secretos en logs.
- Proteger especialmente el entorno `production`.
- Confirmar antes de copiar o descargar todos los secretos.
- Mantener `API_KEY` y las credenciales de Auth0 fuera de la UI pública.
- Registrar quién exportó secretos y cuándo.

## Migración

El `Project` actual agrupa todos y es plano. No debe reutilizarse directamente sin migración.

La nueva estructura debe convertir:

- El proyecto actual `homelab` en `Product`.
- Crear subproyectos como `backend`, `frontend`, `cli` e `infra`.
- Conservar las asociaciones de los todos existentes.

## Fuera De Alcance Inicial

- Inyección automática de secretos en Docker.
- Rotación automática.
- Integración con un secret manager externo.
- Permisos avanzados por equipo.

## Criterios De Aceptación

- Un usuario puede seleccionar producto, proyecto y entorno.
- Puede copiar todo el entorno con una sola acción.
- Puede descargarlo como `.env`.
- Los filtros de búsqueda no limitan la exportación.
- Los valores se mantienen ocultos hasta una acción explícita.
- Los tests cubren formato, permisos y exportación completa.
