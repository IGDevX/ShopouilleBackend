db = db.getSiblingDB('shopouille-cms');

db.createUser({
  user: 'shopouille',
  pwd: 'password',
  roles: [
    {
      role: 'readWrite',
      db: 'shopouille-cms'
    }
  ]
});

db.createCollection('theme-settings', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      properties: {
        primaryColor: {
          bsonType: 'string',
          description: 'Primary color of the theme'
        },
        secondaryColor: {
          bsonType: 'string',
          description: 'Secondary color of the theme'
        }
      }
    }
  }
});

db['theme-settings'].insertOne({
  primaryColor: '#007bff',
  secondaryColor: '#6c757d'
});

db.createCollection('page-content', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      properties: {
        type: {
          bsonType: 'string',
          description: 'Page content type'
        },
        title: {
          bsonType: 'string',
          description: 'Page title'
        },
        content: {
          bsonType: 'string',
          description: 'Page content'
        }
      }
    }
  }
});

db['page-content'].insertOne({
  type: 'home',
  title: 'Welcome',
  content: 'Welcome to our store'
});

db['page-content'].insertOne({
  type: 'cgv',
  title: 'Conditions Générales de Vente',
  content: 'Conditions générales de vente'
});

db['page-content'].insertOne({
  type: 'contact',
  title: 'Contact',
  content: 'Contactez-nous'
});
