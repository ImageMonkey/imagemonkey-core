// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add('login', (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add('drag', { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add('dismiss', { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This will overwrite an existing command --
// Cypress.Commands.overwrite('visit', (originalFn, url, options) => { ... })

import 'cypress-file-upload';

Cypress.Commands.add('clear_db', () => {
    cy.exec('cd ../ && go test -run TestDatabaseEmpty');
});

Cypress.Commands.add('donate_image', (imageName) => {
    cy.visit('http://127.0.0.1:8080/donate');

	cy.fixture('images/apples/'+imageName).then(fileContent => {
        cy.get('[id="dropzone"]').attachFile({
            fileContent: fileContent.toString(),
            fileName: imageName,
            mimeType: 'image/jpeg'
        });
    });
});

Cypress.Commands.add('signup_user', (username, password, emailAddress) => {
	cy.visit('http://127.0.0.1:8080/signup');
	cy.get('#usernameInput').type(username);
	cy.get('#passwordInput').type(password);
	cy.get('#repeatedPasswordInput').type(password);
	cy.get('#emailInput').type(emailAddress);

	cy.get('#signupButton').click();
});
