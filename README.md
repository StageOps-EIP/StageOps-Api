# StageOps API

Backend central de la plateforme StageOps.

Cette API gère la logique métier, la persistance des données et les permissions utilisateurs pour l’ensemble du système.

## Responsabilités

- Authentification utilisateurs
- Gestion des rôles techniques
- Gestion des projets
- Inventaire matériel
- Suivi de l’état opérationnel
- Déclaration d’incidents
- Permissions par module
- Exposition API REST

## Architecture

Architecture modulaire orientée domaine.

src/
  modules/
    auth/
    users/
    projects/
    roles/
    lighting/
    sound/
    equipment/
    incidents/
  database/
  config/
  middleware/

## Stack technique

- Node.js
- Express / NestJS (selon implémentation)
- PostgreSQL
- JWT Authentication
- Validation middleware

## Installation

### Prérequis
- Node.js >= 18
- PostgreSQL
- npm ou yarn

### Setup

git clone <repo>
cd StageOps-Api
npm install

Créer un fichier `.env`

Variables requises :

DATABASE_URL=
JWT_SECRET=
PORT=3000

### Lancement

npm run dev

API accessible sur :
http://localhost:3000

## Convention API

- RESTful endpoints
- JSON uniquement
- Auth Bearer Token

## Sécurité

- Validation des entrées
- Permissions par rôle
- Séparation des modules

## Tests

npm run test

## Objectif MVP

Fournir une API stable pour :

- projets
- matériel
- incidents
- rôles

