import React from 'react';
import {Modal, Button} from '@patternfly/react-core';

const ConfirmDelete = (onClose) => {
  const [isModalOpen, setIsModalOpen] = React.useState(false);
  const onConfirm = () => {
    onClose && onClose(true)
    setIsModalOpen(false)
  }
  const onCancel = () => {
    onClose && onClose(false)
    setIsModalOpen(false)
  }
  return {
    async open() {
      setIsModalOpen(true)
    },
    render() {
      return (
        <Modal
          isSmall
          showClose={false}
          isOpen={isModalOpen}
          title="Please Confirm Deletion"
          actions={[
            <Button key="confirm" variant="primary" onClick={onConfirm}>Delete</Button>,
            <Button key="cancel" variant="link" onClick={onCancel}>Cancel</Button>
          ]}
        >
          Are you sure?
        </Modal>
      );
    }
  }
}
export default ConfirmDelete