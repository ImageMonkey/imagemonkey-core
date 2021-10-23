// npm install --save-dev cypress-file-upload
// start with: node_modules/.bin/cypress open --project tests/ui/

describe('Donate Image', () => {
    before(() => {
        cy.clear_db();
    })

    it('Testing image donation', () => {
        cy.donate_image('apple1.jpeg');
    });
})