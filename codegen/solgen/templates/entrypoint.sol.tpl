// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

{{ range .Imports }}
import "{{ . }}";
{{- end }}

abstract contract {{$.Name}} is {{ range $i, $v := .Interfaces }}{{ if $i }}, {{ end }}{{ $v }}{{ end }} {
    function {{$.ArchParams.MultiActionMethodName}}(
        uint32[] memory actionIds,
        uint8[] memory actionCount,
        bytes[] memory actionData
    ) external {
        uint256 actionIdx = 0;
        for (uint256 i = 0; i < actionIds.length; i++) {
            uint256 numActions = uint256(actionCount[i]);
            for (uint256 j = actionIdx; j < actionIdx + numActions; j++) {
                _executeAction(actionIds[i], actionData[j]);
            }
            actionIdx += numActions;
        }
    }

    function _executeAction(uint32 actionId, bytes memory actionData) private {
        if (actionId == {{$.ArchParams.TickActionIdHex}}) {
            {{ SolidityActionMethodNameFn .ArchParams.TickActionName }}();
        } else if (actionId == {{$.ArchParams.PurgeActionIdHex}}) {
            {{ SolidityActionMethodNameFn .ArchParams.PurgeActionName }}();
        } else 
        {{- range $schema := .Schemas }}
        {{- if $schema.Values }}
        if (actionId == {{ _actionId $schema }}) {
            {{ SolidityActionStructNameFn $schema.Name }} memory action = abi.decode(
                actionData,
                ({{ SolidityActionStructNameFn $schema.Name }})
            );
            {{ SolidityActionMethodNameFn $schema.Name }}(action);
        }
        {{- else }}
        if (actionId == {{ _actionId $schema }}) {
            {{ SolidityActionMethodNameFn $schema.Name }}();
        }
        {{- end}} else {{- end }} {
            revert("Entrypoint: Invalid action ID");
        }
    }

    function {{ SolidityActionMethodNameFn $.ArchParams.TickActionName }}() public virtual {
        revert("not implemented");
    }

    function {{ SolidityActionMethodNameFn $.ArchParams.PurgeActionName }}() public virtual {
        revert("not implemented");
    }

    {{- range $schema := .Schemas }}
    {{ if $schema.Values }}
    function {{ SolidityActionMethodNameFn .Name }}({{ SolidityActionStructNameFn $schema.Name }} memory action) public virtual {
    {{- else }}
    function {{ SolidityActionMethodNameFn .Name }}() public virtual {
    {{- end }}
        revert("not implemented");
    }
    {{- end }}
}
