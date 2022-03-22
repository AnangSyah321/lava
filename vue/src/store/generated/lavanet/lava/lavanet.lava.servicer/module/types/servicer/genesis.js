/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import { Params } from "../servicer/params";
import { StakeMap } from "../servicer/stake_map";
import { SpecStakeStorage } from "../servicer/spec_stake_storage";
import { BlockDeadlineForCallback } from "../servicer/block_deadline_for_callback";
import { UnstakingServicersAllSpecs } from "../servicer/unstaking_servicers_all_specs";
export const protobufPackage = "lavanet.lava.servicer";
const baseGenesisState = { unstakingServicersAllSpecsCount: 0 };
export const GenesisState = {
    encode(message, writer = Writer.create()) {
        if (message.params !== undefined) {
            Params.encode(message.params, writer.uint32(10).fork()).ldelim();
        }
        for (const v of message.stakeMapList) {
            StakeMap.encode(v, writer.uint32(18).fork()).ldelim();
        }
        for (const v of message.specStakeStorageList) {
            SpecStakeStorage.encode(v, writer.uint32(26).fork()).ldelim();
        }
        if (message.blockDeadlineForCallback !== undefined) {
            BlockDeadlineForCallback.encode(message.blockDeadlineForCallback, writer.uint32(34).fork()).ldelim();
        }
        for (const v of message.unstakingServicersAllSpecsList) {
            UnstakingServicersAllSpecs.encode(v, writer.uint32(42).fork()).ldelim();
        }
        if (message.unstakingServicersAllSpecsCount !== 0) {
            writer.uint32(48).uint64(message.unstakingServicersAllSpecsCount);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseGenesisState };
        message.stakeMapList = [];
        message.specStakeStorageList = [];
        message.unstakingServicersAllSpecsList = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.params = Params.decode(reader, reader.uint32());
                    break;
                case 2:
                    message.stakeMapList.push(StakeMap.decode(reader, reader.uint32()));
                    break;
                case 3:
                    message.specStakeStorageList.push(SpecStakeStorage.decode(reader, reader.uint32()));
                    break;
                case 4:
                    message.blockDeadlineForCallback = BlockDeadlineForCallback.decode(reader, reader.uint32());
                    break;
                case 5:
                    message.unstakingServicersAllSpecsList.push(UnstakingServicersAllSpecs.decode(reader, reader.uint32()));
                    break;
                case 6:
                    message.unstakingServicersAllSpecsCount = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseGenesisState };
        message.stakeMapList = [];
        message.specStakeStorageList = [];
        message.unstakingServicersAllSpecsList = [];
        if (object.params !== undefined && object.params !== null) {
            message.params = Params.fromJSON(object.params);
        }
        else {
            message.params = undefined;
        }
        if (object.stakeMapList !== undefined && object.stakeMapList !== null) {
            for (const e of object.stakeMapList) {
                message.stakeMapList.push(StakeMap.fromJSON(e));
            }
        }
        if (object.specStakeStorageList !== undefined &&
            object.specStakeStorageList !== null) {
            for (const e of object.specStakeStorageList) {
                message.specStakeStorageList.push(SpecStakeStorage.fromJSON(e));
            }
        }
        if (object.blockDeadlineForCallback !== undefined &&
            object.blockDeadlineForCallback !== null) {
            message.blockDeadlineForCallback = BlockDeadlineForCallback.fromJSON(object.blockDeadlineForCallback);
        }
        else {
            message.blockDeadlineForCallback = undefined;
        }
        if (object.unstakingServicersAllSpecsList !== undefined &&
            object.unstakingServicersAllSpecsList !== null) {
            for (const e of object.unstakingServicersAllSpecsList) {
                message.unstakingServicersAllSpecsList.push(UnstakingServicersAllSpecs.fromJSON(e));
            }
        }
        if (object.unstakingServicersAllSpecsCount !== undefined &&
            object.unstakingServicersAllSpecsCount !== null) {
            message.unstakingServicersAllSpecsCount = Number(object.unstakingServicersAllSpecsCount);
        }
        else {
            message.unstakingServicersAllSpecsCount = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.params !== undefined &&
            (obj.params = message.params ? Params.toJSON(message.params) : undefined);
        if (message.stakeMapList) {
            obj.stakeMapList = message.stakeMapList.map((e) => e ? StakeMap.toJSON(e) : undefined);
        }
        else {
            obj.stakeMapList = [];
        }
        if (message.specStakeStorageList) {
            obj.specStakeStorageList = message.specStakeStorageList.map((e) => e ? SpecStakeStorage.toJSON(e) : undefined);
        }
        else {
            obj.specStakeStorageList = [];
        }
        message.blockDeadlineForCallback !== undefined &&
            (obj.blockDeadlineForCallback = message.blockDeadlineForCallback
                ? BlockDeadlineForCallback.toJSON(message.blockDeadlineForCallback)
                : undefined);
        if (message.unstakingServicersAllSpecsList) {
            obj.unstakingServicersAllSpecsList = message.unstakingServicersAllSpecsList.map((e) => (e ? UnstakingServicersAllSpecs.toJSON(e) : undefined));
        }
        else {
            obj.unstakingServicersAllSpecsList = [];
        }
        message.unstakingServicersAllSpecsCount !== undefined &&
            (obj.unstakingServicersAllSpecsCount =
                message.unstakingServicersAllSpecsCount);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseGenesisState };
        message.stakeMapList = [];
        message.specStakeStorageList = [];
        message.unstakingServicersAllSpecsList = [];
        if (object.params !== undefined && object.params !== null) {
            message.params = Params.fromPartial(object.params);
        }
        else {
            message.params = undefined;
        }
        if (object.stakeMapList !== undefined && object.stakeMapList !== null) {
            for (const e of object.stakeMapList) {
                message.stakeMapList.push(StakeMap.fromPartial(e));
            }
        }
        if (object.specStakeStorageList !== undefined &&
            object.specStakeStorageList !== null) {
            for (const e of object.specStakeStorageList) {
                message.specStakeStorageList.push(SpecStakeStorage.fromPartial(e));
            }
        }
        if (object.blockDeadlineForCallback !== undefined &&
            object.blockDeadlineForCallback !== null) {
            message.blockDeadlineForCallback = BlockDeadlineForCallback.fromPartial(object.blockDeadlineForCallback);
        }
        else {
            message.blockDeadlineForCallback = undefined;
        }
        if (object.unstakingServicersAllSpecsList !== undefined &&
            object.unstakingServicersAllSpecsList !== null) {
            for (const e of object.unstakingServicersAllSpecsList) {
                message.unstakingServicersAllSpecsList.push(UnstakingServicersAllSpecs.fromPartial(e));
            }
        }
        if (object.unstakingServicersAllSpecsCount !== undefined &&
            object.unstakingServicersAllSpecsCount !== null) {
            message.unstakingServicersAllSpecsCount =
                object.unstakingServicersAllSpecsCount;
        }
        else {
            message.unstakingServicersAllSpecsCount = 0;
        }
        return message;
    },
};
var globalThis = (() => {
    if (typeof globalThis !== "undefined")
        return globalThis;
    if (typeof self !== "undefined")
        return self;
    if (typeof window !== "undefined")
        return window;
    if (typeof global !== "undefined")
        return global;
    throw "Unable to locate global object";
})();
function longToNumber(long) {
    if (long.gt(Number.MAX_SAFE_INTEGER)) {
        throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
    }
    return long.toNumber();
}
if (util.Long !== Long) {
    util.Long = Long;
    configure();
}