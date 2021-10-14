describe('Unified Mode', () => {
    before(() => {
        cy.clear_db_and_create_moderator_account();
        cy.donate_image('apple1.jpeg');
        cy.donate_image('apple2.jpeg');
        cy.login('moderator', 'moderator');
        cy.unlock_all_images();
    })

    it('Browse Image', () => {
        cy.query_images("image.unlabeled='true'", 2);
    });

    it('Open Image for Annotation', () => {
        cy.query_images("image.unlabeled='true'", 2);
        cy.get('#annotation-image-grid').find('img').first().then((elem) => {
            elem.click();
            cy.get('#loading-spinner').should('not.be.visible');
        });
    });

    it('Open Image for Annotation', () => {
        cy.query_images("image.unlabeled='true'", 2);
        cy.get('#annotation-image-grid').find('img').first().then((elem) => {
            elem.click();
        });
    });
})