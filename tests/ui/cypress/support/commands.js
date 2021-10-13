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
import 'cypress-xpath';

Cypress.Commands.add('clear_db', () => {
    cy.exec('cd ../ && go test -run TestDatabaseEmpty');
});

Cypress.Commands.add('clear_db_and_create_moderator_account', () => {
    cy.exec('cd ../ && go test -run TestDatabaseEmptyWithUserThatHasUnlockImagePermission');
});

Cypress.Commands.add('donate_image', (imageName) => {
    cy.visit('http://127.0.0.1:8080/donate');

    //cy.wait(500);

    cy.fixture('images/apples/' + imageName).then(fileContent => {
        console.log(fileContent.toString());
        cy.xpath("//input[@type='file']").attachFile({
            fileContent: fileContent.toString(),
            filePath: 'images/apples/' + imageName
        });
    });

    //wait until image is uploaded or at max 10 seconds
    cy.get('#successMsg', {
        timeout: 10000
    }).should('be.visible');
});

Cypress.Commands.add('signup_user', (username, password, emailAddress) => {
    cy.visit('http://127.0.0.1:8080/signup');
    cy.get('#usernameInput').type(username);
    cy.get('#passwordInput').type(password);
    cy.get('#repeatedPasswordInput').type(password);
    cy.get('#emailInput').type(emailAddress);

    cy.get('#signupButton').click();
});

Cypress.Commands.add('query_images', (query) => {
    cy.visit('http://127.0.0.1:8080/annotate?mode=browse&view=unified&v=2');
});

Cypress.Commands.add('unlock_all_images', () => {
    cy.visit('http://127.0.0.1:8080/image_unlock?mode=browse');

    cy.get('[id^=galleryitem]').click({
        multiple: true,
        force: true
    });
    cy.get('#imageUnlockDoneButton').click();
});

Cypress.Commands.add('login', (username, password) => {
    cy.visit('http://127.0.0.1:8080/login');

    cy.get('#usernameInput').type(username);
    cy.get('#passwordInput').type(password);
    cy.get('#loginButton').click();

    //wait until redirected after login or at max 4 seconds
    cy.url().should('eq', 'http://127.0.0.1:8080/')
});